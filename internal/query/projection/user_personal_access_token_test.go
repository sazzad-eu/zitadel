package projection

import (
	"testing"
	"time"

	"github.com/zitadel/zitadel/internal/database"
	"github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/eventstore/handler"
	"github.com/zitadel/zitadel/internal/eventstore/repository"
	"github.com/zitadel/zitadel/internal/repository/instance"
	"github.com/zitadel/zitadel/internal/repository/org"
	"github.com/zitadel/zitadel/internal/repository/user"
)

func TestPersonalAccessTokenProjection_reduces(t *testing.T) {
	type args struct {
		event func(t *testing.T) eventstore.Event
	}
	tests := []struct {
		name   string
		args   args
		reduce func(event eventstore.Event) (*handler.Statement, error)
		want   wantReduce
	}{
		{
			name: "reducePersonalAccessTokenAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.PersonalAccessTokenAddedType),
					user.AggregateType,
					[]byte(`{"tokenId": "tokenID", "expiration": "9999-12-31T23:59:59Z", "scopes": ["openid"]}`),
				), user.PersonalAccessTokenAddedEventMapper),
			},
			reduce: (&personalAccessTokenProjection{}).reducePersonalAccessTokenAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("user"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.personal_access_tokens3 (id, creation_date, change_date, resource_owner, instance_id, sequence, user_id, expiration, scopes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
							expectedArgs: []interface{}{
								"tokenID",
								anyArg{},
								anyArg{},
								"ro-id",
								"instance-id",
								uint64(15),
								"agg-id",
								time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC),
								database.StringArray{"openid"},
							},
						},
					},
				},
			},
		},
		{
			name: "reducePersonalAccessTokenRemoved",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.PersonalAccessTokenRemovedType),
					user.AggregateType,
					[]byte(`{"tokenId": "tokenID"}`),
				), user.PersonalAccessTokenRemovedEventMapper),
			},
			reduce: (&personalAccessTokenProjection{}).reducePersonalAccessTokenRemoved,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("user"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.personal_access_tokens3 WHERE (id = $1) AND (instance_id = $2)",
							expectedArgs: []interface{}{
								"tokenID",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "reduceUserRemoved",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.PersonalAccessTokenRemovedType),
					user.AggregateType,
					nil,
				), user.UserRemovedEventMapper),
			},
			reduce: (&personalAccessTokenProjection{}).reduceUserRemoved,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("user"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.personal_access_tokens3 WHERE (user_id = $1) AND (instance_id = $2)",
							expectedArgs: []interface{}{
								"agg-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name:   "org reduceOwnerRemoved",
			reduce: (&personalAccessTokenProjection{}).reduceOwnerRemoved,
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.OrgRemovedEventType),
					org.AggregateType,
					nil,
				), org.OrgRemovedEventMapper),
			},
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.personal_access_tokens3 SET (change_date, sequence, owner_removed) = ($1, $2, $3) WHERE (instance_id = $4) AND (resource_owner = $5)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								true,
								"instance-id",
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceInstanceRemoved",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.InstanceRemovedEventType),
					instance.AggregateType,
					nil,
				), instance.InstanceRemovedEventMapper),
			},
			reduce: reduceInstanceRemovedHelper(PersonalAccessTokenColumnInstanceID),
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.personal_access_tokens3 WHERE (instance_id = $1)",
							expectedArgs: []interface{}{
								"agg-id",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := baseEvent(t)
			got, err := tt.reduce(event)
			if _, ok := err.(errors.InvalidArgument); !ok {
				t.Errorf("no wrong event mapping: %v, got: %v", err, got)
			}

			event = tt.args.event(t)
			got, err = tt.reduce(event)
			assertReduce(t, got, err, PersonalAccessTokenProjectionTable, tt.want)
		})
	}
}
