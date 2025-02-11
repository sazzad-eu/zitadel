package command

import (
	"time"

	"github.com/zitadel/zitadel/internal/eventstore"

	"github.com/zitadel/zitadel/internal/crypto"
	"github.com/zitadel/zitadel/internal/domain"
	"github.com/zitadel/zitadel/internal/repository/user"
)

type HumanPasswordWriteModel struct {
	eventstore.WriteModel

	EncodedHash          string
	SecretChangeRequired bool

	Code                     *crypto.CryptoValue
	CodeCreationDate         time.Time
	CodeExpiry               time.Duration
	PasswordCheckFailedCount uint64

	UserState domain.UserState
}

func NewHumanPasswordWriteModel(userID, resourceOwner string) *HumanPasswordWriteModel {
	return &HumanPasswordWriteModel{
		WriteModel: eventstore.WriteModel{
			AggregateID:   userID,
			ResourceOwner: resourceOwner,
		},
	}
}

func (wm *HumanPasswordWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *user.HumanAddedEvent:
			wm.EncodedHash = user.SecretOrEncodedHash(e.Secret, e.EncodedHash)
			wm.SecretChangeRequired = e.ChangeRequired
			wm.UserState = domain.UserStateActive
		case *user.HumanRegisteredEvent:
			wm.EncodedHash = user.SecretOrEncodedHash(e.Secret, e.EncodedHash)
			wm.SecretChangeRequired = e.ChangeRequired
			wm.UserState = domain.UserStateActive
		case *user.HumanInitialCodeAddedEvent:
			wm.UserState = domain.UserStateInitial
		case *user.HumanInitializedCheckSucceededEvent:
			wm.UserState = domain.UserStateActive
		case *user.HumanPasswordChangedEvent:
			wm.EncodedHash = user.SecretOrEncodedHash(e.Secret, e.EncodedHash)
			wm.SecretChangeRequired = e.ChangeRequired
			wm.Code = nil
			wm.PasswordCheckFailedCount = 0
		case *user.HumanPasswordCodeAddedEvent:
			wm.Code = e.Code
			wm.CodeCreationDate = e.CreationDate()
			wm.CodeExpiry = e.Expiry
		case *user.HumanEmailVerifiedEvent:
			if wm.UserState == domain.UserStateInitial {
				wm.UserState = domain.UserStateActive
			}
		case *user.HumanPasswordCheckFailedEvent:
			wm.PasswordCheckFailedCount += 1
		case *user.HumanPasswordCheckSucceededEvent:
			wm.PasswordCheckFailedCount = 0
		case *user.UserUnlockedEvent:
			wm.PasswordCheckFailedCount = 0
		case *user.UserRemovedEvent:
			wm.UserState = domain.UserStateDeleted
		case *user.HumanPasswordHashUpdatedEvent:
			wm.EncodedHash = e.EncodedHash
		}
	}
	return wm.WriteModel.Reduce()
}

func (wm *HumanPasswordWriteModel) Query() *eventstore.SearchQueryBuilder {
	query := eventstore.NewSearchQueryBuilder(eventstore.ColumnsEvent).
		AddQuery().
		AggregateTypes(user.AggregateType).
		AggregateIDs(wm.AggregateID).
		EventTypes(user.HumanAddedType,
			user.HumanRegisteredType,
			user.HumanInitialCodeAddedType,
			user.HumanInitializedCheckSucceededType,
			user.HumanPasswordChangedType,
			user.HumanPasswordCodeAddedType,
			user.HumanEmailVerifiedType,
			user.HumanPasswordCheckFailedType,
			user.HumanPasswordCheckSucceededType,
			user.UserRemovedType,
			user.UserUnlockedType,
			user.UserV1AddedType,
			user.UserV1RegisteredType,
			user.UserV1InitialCodeAddedType,
			user.UserV1InitializedCheckSucceededType,
			user.UserV1PasswordChangedType,
			user.UserV1PasswordCodeAddedType,
			user.UserV1EmailVerifiedType,
			user.UserV1PasswordCheckFailedType,
			user.UserV1PasswordCheckSucceededType,
			user.UserV1PasswordHashUpdatedType).
		Builder()

	if wm.ResourceOwner != "" {
		query.ResourceOwner(wm.ResourceOwner)
	}
	return query
}
