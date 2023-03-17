CREATE MATERIALIZED VIEW hyperparameters_view AS
    SELECT workspaces.id AS workspace_id,  experiments.project_id AS project_id, experiments.id AS experiment_id, json_build_array((experiments.config->'hyperparameters')) AS hyperparameters FROM workspaces, experiments, projects WHERE workspaces.id = projects.workspace_id AND experiments.project_id = projects.id;

CREATE INDEX ix_hyperparameters_workspace_id ON hyperparameters_view USING btree (workspace_id);
CREATE INDEX ix_hyperparameters_project_id ON hyperparameters_view USING btree (project_id);


CREATE MATERIALIZED VIEW metrics_name_view AS  
    WITH training_metrics_names AS (
        SELECT array_to_json(array_agg(DISTINCT names)) AS name, e.id AS experiment_id
        FROM trials t, experiments e, steps s,
            LATERAL jsonb_object_keys(s.metrics->'avg_metrics') AS names
        WHERE t.id=s.trial_id AND e.id = t.experiment_id 
        GROUP BY e.id),
    validation_metrics_names AS (
        SELECT array_to_json(array_agg(DISTINCT names)) AS name, e.id AS experiment_id
        FROM trials t, experiments e, validations v,
            LATERAL jsonb_object_keys(v.metrics->'validation_metrics') AS names
        WHERE t.id=v.trial_id AND e.id = t.experiment_id 
        GROUP BY e.id)
    SELECT 
        workspaces.id AS workspace_id,  
        e.project_id AS project_id, 
        e.id AS experiment_id, 
        COALESCE(training_metrics_names.name, '[]'::json) AS tname,
        COALESCE(validation_metrics_names.name, '[]'::json) AS vname 
    FROM workspaces, experiments e LEFT JOIN validation_metrics_names ON e.id = validation_metrics_names.experiment_id, projects, training_metrics_names 
    WHERE 
        workspaces.id = projects.workspace_id AND 
        e.project_id = projects.id AND 
        training_metrics_names.experiment_id = e.id;


CREATE INDEX ix_metrics_name_workspace_id ON metrics_name_view USING btree (workspace_id);
CREATE INDEX ix_metrics_name_project_id ON metrics_name_view USING btree (project_id);