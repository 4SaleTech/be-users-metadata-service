package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/be-users-metadata-service/internal/application/ports"
	"github.com/be-users-metadata-service/internal/application/service"
	"github.com/be-users-metadata-service/internal/domain"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var errMissingUserID = errors.New("missing user_id")

// ProcessEvent runs the process-event use case: idempotency check, load rules, evaluate, update meta or record processed/failed.
type ProcessEvent struct {
	ruleRepo         ports.RuleRepository
	userRepo         ports.UserRepository
	failedEventRepo  ports.FailedEventRepository
	ruleEngine       *service.RuleEngine
	metadataExecutor *service.MetadataExecutor
	idempotency      *service.IdempotencyService
	log              *slog.Logger
}

// NewProcessEvent returns a new ProcessEvent use case.
func NewProcessEvent(
	ruleRepo ports.RuleRepository,
	userRepo ports.UserRepository,
	failedEventRepo ports.FailedEventRepository,
	ruleEngine *service.RuleEngine,
	metadataExecutor *service.MetadataExecutor,
	idempotency *service.IdempotencyService,
	log *slog.Logger,
) *ProcessEvent {
	return &ProcessEvent{
		ruleRepo:         ruleRepo,
		userRepo:         userRepo,
		failedEventRepo:  failedEventRepo,
		ruleEngine:       ruleEngine,
		metadataExecutor: metadataExecutor,
		idempotency:      idempotency,
		log:              log,
	}
}

// ProcessEvent handles one event (idempotent). Returns nil to ack, or error to nack.
func (u *ProcessEvent) ProcessEvent(ctx context.Context, event domain.Event, rawPayload []byte) error {
	eventID := idempotencyKey(&event, rawPayload)
	log := u.log.With("event_id", eventID, "event_type", event.Type)

	log.Debug("checking idempotency")
	alreadyProcessed, err := u.idempotency.Exists(ctx, eventID)
	if err != nil {
		log.Error("idempotency check failed", "error", err)
		return err
	}
	if alreadyProcessed {
		log.Info("event already processed, skipping")
		return nil
	}

	userID, err := u.resolveUserID(event)
	if err != nil {
		log.Warn("missing user_id, recording failed event", "reason", err.Error())
		_ = u.failedEventRepo.CreateFromError(ctx, event.Type, rawPayload, err.Error())
		return nil
	}
	log.Debug("loaded user", "user_id", userID)

	rules, err := u.ruleRepo.FindByEventTypeAndVersion(ctx, event.Type, event.Version)
	if err != nil {
		log.Error("load rules failed", "error", err)
		return err
	}
	log.Debug("rules loaded", "count", len(rules))
	if len(rules) == 0 {
		log.Warn("no rules for event type, recording failed event")
		_ = u.failedEventRepo.CreateFromError(ctx, event.Type, rawPayload, "no rules for event type")
		return nil
	}

	currentMeta, metaMap, err := u.getUserMetaAndMap(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("user not found, recording failed event")
			_ = u.failedEventRepo.CreateFromError(ctx, event.Type, rawPayload, "user not found")
			return nil
		}
		log.Error("fetch metadata failed", "error", err)
		return err
	}

	operations, err := u.ruleEngine.EvaluateRules(ctx, &event, rules, metaMap)
	if err != nil {
		log.Error("evaluate rules failed", "error", err)
		return err
	}
	log.Debug("rules evaluated", "operations", len(operations))
	if len(operations) == 0 {
		log.Info("no operations applied, recording processed only")
		if err := u.idempotency.RecordProcessedOnly(ctx, eventID, rawPayload); err != nil {
			log.Error("record processed only failed", "error", err)
			return err
		}
		log.Info("event processed successfully", "metadata_updated", false)
		return nil
	}

	newMeta, err := u.metadataExecutor.ComputeResult(ctx, currentMeta, operations)
	if err != nil {
		log.Error("compute metadata failed", "error", err)
		return err
	}

	if err := u.idempotency.RecordSuccess(ctx, userID, newMeta, eventID, rawPayload); err != nil {
		log.Error("record success failed", "error", err)
		return err
	}
	log.Info("event processed successfully", "metadata_updated", true)
	return nil
}

func (u *ProcessEvent) resolveUserID(event domain.Event) (string, error) {
	if strings.TrimSpace(event.UserID) == "" {
		return "", errMissingUserID
	}
	return event.UserID, nil
}

// idempotencyKey returns event.ID if set; otherwise a deterministic hash of the payload so same event = same key.
func idempotencyKey(event *domain.Event, rawPayload []byte) string {
	if event != nil && strings.TrimSpace(event.ID) != "" {
		return event.ID
	}
	if len(rawPayload) == 0 {
		return uuid.New().String() // no payload to hash, fallback
	}
	h := sha256.Sum256(rawPayload)
	return hex.EncodeToString(h[:])
}

func (u *ProcessEvent) getUserMetaAndMap(ctx context.Context, userID string) (datatypes.JSON, map[string]interface{}, error) {
	meta, err := u.userRepo.GetMetaData(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if len(meta) == 0 {
		meta = datatypes.JSON([]byte("{}"))
	}
	var m map[string]interface{}
	if err := json.Unmarshal(meta, &m); err != nil {
		m = make(map[string]interface{})
	}
	if m == nil {
		m = make(map[string]interface{})
	}
	return meta, m, nil
}
