package ngalert

import (
	"context"

	"github.com/grafana/grafana/pkg/services/sqlstore"
)

func getAlertDefinitionByID(alertDefinitionID int64, sess *sqlstore.DBSession) (*AlertDefinition, error) {
	alertDefinition := AlertDefinition{}
	has, err := sess.ID(alertDefinitionID).Get(&alertDefinition)
	if !has {
		return nil, errAlertDefinitionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &alertDefinition, nil
}

// deleteAlertDefinitionByID is a handler for deleting an alert definition.
// It returns models.ErrAlertDefinitionNotFound if no alert definition is found for the provided ID.
func (ng *AlertNG) deleteAlertDefinitionByID(cmd *deleteAlertDefinitionByIDCommand) error {
	return ng.SQLStore.WithTransactionalDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		res, err := sess.Exec("DELETE FROM alert_definition WHERE id = ?", cmd.ID)
		if err != nil {
			return err
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		cmd.RowsAffected = rowsAffected
		return nil
	})
}

// getAlertDefinitionByID is a handler for retrieving an alert definition from that database by its ID.
// It returns models.ErrAlertDefinitionNotFound if no alert definition is found for the provided ID.
func (ng *AlertNG) getAlertDefinitionByID(query *getAlertDefinitionByIDQuery) error {
	return ng.SQLStore.WithTransactionalDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		alertDefinition, err := getAlertDefinitionByID(query.ID, sess)
		if err != nil {
			return err
		}
		query.Result = alertDefinition
		return nil
	})
}

// saveAlertDefinition is a handler for saving a new alert definition.
func (ng *AlertNG) saveAlertDefinition(cmd *saveAlertDefinitionCommand) error {
	return ng.SQLStore.WithTransactionalDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		intervalInSeconds := defaultIntervalInSeconds
		if cmd.IntervalInSeconds != nil {
			intervalInSeconds = *cmd.IntervalInSeconds
		}
		alertDefinition := &AlertDefinition{
			OrgId:     cmd.OrgID,
			Name:      cmd.Name,
			Condition: cmd.Condition.RefID,
			Data:      cmd.Condition.QueriesAndExpressions,
			Interval:  intervalInSeconds,
		}

		if err := ng.validateAlertDefinition(alertDefinition, false); err != nil {
			return err
		}

		if err := alertDefinition.preSave(); err != nil {
			return err
		}

		if _, err := sess.Insert(alertDefinition); err != nil {
			return err
		}

		cmd.Result = alertDefinition
		return nil
	})
}

// updateAlertDefinition is a handler for updating an existing alert definition.
// It returns models.ErrAlertDefinitionNotFound if no alert definition is found for the provided ID.
func (ng *AlertNG) updateAlertDefinition(cmd *updateAlertDefinitionCommand) error {
	return ng.SQLStore.WithTransactionalDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		alertDefinition := &AlertDefinition{
			Id:        cmd.ID,
			Name:      cmd.Name,
			Condition: cmd.Condition.RefID,
			Data:      cmd.Condition.QueriesAndExpressions,
		}
		if cmd.IntervalInSeconds != nil {
			alertDefinition.Interval = *cmd.IntervalInSeconds
		}

		if err := ng.validateAlertDefinition(alertDefinition, true); err != nil {
			return err
		}

		if err := alertDefinition.preSave(); err != nil {
			return err
		}

		affectedRows, err := sess.ID(cmd.ID).Update(alertDefinition)
		if err != nil {
			return err
		}

		cmd.Result = alertDefinition
		cmd.RowsAffected = affectedRows
		return nil
	})
}

// getOrgAlertDefinitions is a handler for retrieving alert definitions of specific organisation.
func (ng *AlertNG) getOrgAlertDefinitions(query *listAlertDefinitionsQuery) error {
	return ng.SQLStore.WithTransactionalDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		alertDefinitions := make([]*AlertDefinition, 0)
		q := "SELECT * FROM alert_definition WHERE org_id = ?"
		if err := sess.SQL(q, query.OrgID).Find(&alertDefinitions); err != nil {
			return err
		}

		query.Result = alertDefinitions
		return nil
	})
}

func (ng *AlertNG) getAlertDefinitions(query *listAlertDefinitionsQuery) error {
	return ng.SQLStore.WithTransactionalDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		alerts := make([]*AlertDefinition, 0)
		q := "SELECT id, interval FROM alert_definition"
		if err := sess.SQL(q).Find(&alerts); err != nil {
			return err
		}

		query.Result = alerts
		return nil
	})
}
