CREATE TABLE recruiters (
    id         uuid        PRIMARY KEY,
    user_id    uuid        NOT NULL,
    name       text        NOT NULL,
    company    text        NOT NULL DEFAULT '',
    phone      text        NOT NULL DEFAULT '',
    email      text        NOT NULL DEFAULT '',
    rating     integer     NOT NULL DEFAULT 0,
    comments   text[]      NOT NULL DEFAULT '{}',
    archived   boolean     NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX recruiters_user_id_idx ON recruiters(user_id);

CREATE TABLE jobs (
    id                  uuid        PRIMARY KEY,
    user_id             uuid        NOT NULL,
    recruiter_id        uuid        NOT NULL REFERENCES recruiters(id),
    job_title           text        NOT NULL,
    work_from           text        NOT NULL DEFAULT '',
    date_applied        text        NOT NULL DEFAULT '',
    company_name        text        NOT NULL,
    company_address     text        NOT NULL DEFAULT '',
    company_city        text        NOT NULL DEFAULT '',
    company_state       text        NOT NULL DEFAULT '',
    point_of_contact    text        NOT NULL DEFAULT '',
    poc_title           text        NOT NULL DEFAULT '',
    interviews          text[]      NOT NULL DEFAULT '{}',
    comments            text[]      NOT NULL DEFAULT '{}',
    status              text        NOT NULL DEFAULT 'applied',
    archived            boolean     NOT NULL DEFAULT false,
    primary_link        text        NOT NULL DEFAULT '',
    primary_link_text   text        NOT NULL DEFAULT '',
    secondary_link      text        NOT NULL DEFAULT '',
    secondary_link_text text        NOT NULL DEFAULT '',
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX jobs_user_id_idx ON jobs(user_id);
CREATE INDEX jobs_recruiter_id_idx ON jobs(recruiter_id);
