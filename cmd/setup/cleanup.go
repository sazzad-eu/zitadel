package setup

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zitadel/logging"

	"github.com/zitadel/zitadel/internal/database"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/migration"
)

func NewCleanup() *cobra.Command {
	return &cobra.Command{
		Use:   "cleanup",
		Short: "cleans up migration if they got stuck",
		Long:  `cleans up migration if they got stuck`,
		Run: func(cmd *cobra.Command, args []string) {
			config := MustNewConfig(viper.GetViper())
			Cleanup(config)
		},
	}
}

func Cleanup(config *Config) {
	ctx := context.Background()

	logging.Info("cleanup started")

	dbClient, err := database.Connect(config.Database, false)
	logging.OnError(err).Fatal("unable to connect to database")

	es, err := eventstore.Start(&eventstore.Config{Client: dbClient})
	logging.OnError(err).Fatal("unable to start eventstore")
	migration.RegisterMappers(es)

	step, err := migration.LatestStep(ctx, es)
	logging.OnError(err).Fatal("unable to query latest migration")

	if step.BaseEvent.EventType != migration.StartedType {
		logging.Info("there is no stuck migration please run `zitadel setup`")
		return
	}

	logging.WithFields("name", step.Name).Info("cleanup migration")

	err = migration.CancelStep(ctx, es, step)
	logging.OnError(err).Fatal("cleanup migration failed please retry")
}