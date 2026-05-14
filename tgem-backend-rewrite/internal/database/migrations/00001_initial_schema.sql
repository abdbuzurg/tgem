-- +goose Up
-- Phase 5 baseline: schema produced by gorm.AutoMigrate against the model
-- definitions in internal/database/database.go as of commit 38a99a2.
-- Captured via: cmd/dump_baseline_schema + pg_dump --schema-only.
-- AutoMigrate is turned off in InitDB by the same commit that introduces
-- this migration; subsequent schema changes are added as new migration
-- files under internal/database/migrations/.
--
-- pg_dump session-config statements (SET statement_timeout, SELECT
-- pg_catalog.set_config('search_path', '', false), SET row_security = off,
-- etc.) are stripped from the dump because they would otherwise mutate
-- the goose runtime transaction (notably emptying the search_path so
-- that goose can't locate its own goose_db_version bookkeeping table).
-- Removing them is safe: they are pg_dump output for psql replay, not
-- schema content.

--
-- PostgreSQL database dump
--


-- Dumped from database version 17.7 (Ubuntu 17.7-0ubuntu0.25.04.1)
-- Dumped by pg_dump version 17.7 (Ubuntu 17.7-0ubuntu0.25.04.1)




--
-- Name: auction_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.auction_items (
    id bigint NOT NULL,
    auction_package_id bigint,
    name text,
    description text,
    unit text,
    quantity numeric,
    note text
);


--
-- Name: auction_items_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.auction_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: auction_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.auction_items_id_seq OWNED BY public.auction_items.id;


--
-- Name: auction_packages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.auction_packages (
    id bigint NOT NULL,
    auction_id bigint,
    name text
);


--
-- Name: auction_packages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.auction_packages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: auction_packages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.auction_packages_id_seq OWNED BY public.auction_packages.id;


--
-- Name: auction_participant_prices; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.auction_participant_prices (
    id bigint NOT NULL,
    auction_item_id bigint,
    user_id bigint,
    unit_price text,
    comments text
);


--
-- Name: auction_participant_prices_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.auction_participant_prices_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: auction_participant_prices_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.auction_participant_prices_id_seq OWNED BY public.auction_participant_prices.id;


--
-- Name: auctions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.auctions (
    id bigint NOT NULL,
    name text
);


--
-- Name: auctions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.auctions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: auctions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.auctions_id_seq OWNED BY public.auctions.id;


--
-- Name: districts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.districts (
    id bigint NOT NULL,
    name text,
    project_id bigint
);


--
-- Name: districts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.districts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: districts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.districts_id_seq OWNED BY public.districts.id;


--
-- Name: invoice_counts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoice_counts (
    id bigint NOT NULL,
    project_id bigint,
    invoice_type text,
    count bigint
);


--
-- Name: invoice_counts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.invoice_counts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: invoice_counts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.invoice_counts_id_seq OWNED BY public.invoice_counts.id;


--
-- Name: invoice_inputs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoice_inputs (
    id bigint NOT NULL,
    project_id bigint,
    warehouse_manager_worker_id bigint,
    released_worker_id bigint,
    delivery_code text,
    notes text,
    date_of_invoice timestamp with time zone,
    confirmed boolean
);


--
-- Name: invoice_inputs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.invoice_inputs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: invoice_inputs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.invoice_inputs_id_seq OWNED BY public.invoice_inputs.id;


--
-- Name: invoice_materials; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoice_materials (
    id bigint NOT NULL,
    project_id bigint,
    material_cost_id bigint,
    invoice_id bigint,
    invoice_type text,
    is_defected boolean,
    amount numeric,
    notes text
);


--
-- Name: invoice_materials_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.invoice_materials_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: invoice_materials_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.invoice_materials_id_seq OWNED BY public.invoice_materials.id;


--
-- Name: invoice_object_operators; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoice_object_operators (
    id bigint NOT NULL,
    operator_worker_id bigint,
    invoice_object_id bigint
);


--
-- Name: invoice_object_operators_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.invoice_object_operators_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: invoice_object_operators_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.invoice_object_operators_id_seq OWNED BY public.invoice_object_operators.id;


--
-- Name: invoice_objects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoice_objects (
    id bigint NOT NULL,
    district_id bigint,
    delivery_code text,
    project_id bigint,
    supervisor_worker_id bigint,
    object_id bigint,
    team_id bigint,
    date_of_invoice timestamp with time zone,
    confirmed_by_operator boolean,
    date_of_correction timestamp with time zone
);


--
-- Name: invoice_objects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.invoice_objects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: invoice_objects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.invoice_objects_id_seq OWNED BY public.invoice_objects.id;


--
-- Name: invoice_operations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoice_operations (
    id bigint NOT NULL,
    project_id bigint,
    operation_id bigint,
    invoice_id bigint,
    invoice_type text,
    amount numeric,
    notes text
);


--
-- Name: invoice_operations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.invoice_operations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: invoice_operations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.invoice_operations_id_seq OWNED BY public.invoice_operations.id;


--
-- Name: invoice_output_out_of_projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoice_output_out_of_projects (
    id bigint NOT NULL,
    project_id bigint,
    delivery_code text,
    released_worker_id bigint,
    name_of_project text,
    date_of_invoice timestamp with time zone,
    notes text,
    confirmation boolean
);


--
-- Name: invoice_output_out_of_projects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.invoice_output_out_of_projects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: invoice_output_out_of_projects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.invoice_output_out_of_projects_id_seq OWNED BY public.invoice_output_out_of_projects.id;


--
-- Name: invoice_outputs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoice_outputs (
    id bigint NOT NULL,
    district_id bigint,
    project_id bigint,
    warehouse_manager_worker_id bigint,
    released_worker_id bigint,
    recipient_worker_id bigint,
    team_id bigint,
    delivery_code text,
    date_of_invoice timestamp with time zone,
    notes text,
    confirmation boolean
);


--
-- Name: invoice_outputs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.invoice_outputs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: invoice_outputs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.invoice_outputs_id_seq OWNED BY public.invoice_outputs.id;


--
-- Name: invoice_returns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoice_returns (
    id bigint NOT NULL,
    project_id bigint,
    district_id bigint,
    returner_type text,
    returner_id bigint,
    acceptor_type text,
    acceptor_id bigint,
    accepted_by_worker_id bigint,
    date_of_invoice timestamp with time zone,
    notes text,
    delivery_code text,
    confirmation boolean
);


--
-- Name: invoice_returns_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.invoice_returns_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: invoice_returns_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.invoice_returns_id_seq OWNED BY public.invoice_returns.id;


--
-- Name: invoice_write_offs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoice_write_offs (
    id bigint NOT NULL,
    project_id bigint,
    released_worker_id bigint,
    write_off_type text,
    write_off_location_id bigint,
    delivery_code text,
    date_of_invoice timestamp with time zone,
    confirmation boolean,
    date_of_confirmation timestamp with time zone,
    notes text
);


--
-- Name: invoice_write_offs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.invoice_write_offs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: invoice_write_offs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.invoice_write_offs_id_seq OWNED BY public.invoice_write_offs.id;


--
-- Name: kl04_kv_objects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.kl04_kv_objects (
    id bigint NOT NULL,
    length numeric,
    nourashes text
);


--
-- Name: kl04_kv_objects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.kl04_kv_objects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: kl04_kv_objects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.kl04_kv_objects_id_seq OWNED BY public.kl04_kv_objects.id;


--
-- Name: material_costs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.material_costs (
    id bigint NOT NULL,
    material_id bigint,
    cost_prime numeric(20,4),
    cost_m19 numeric(20,4),
    cost_with_customer numeric(20,4)
);


--
-- Name: material_costs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.material_costs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: material_costs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.material_costs_id_seq OWNED BY public.material_costs.id;


--
-- Name: material_defects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.material_defects (
    id bigint NOT NULL,
    amount numeric,
    material_location_id bigint
);


--
-- Name: material_defects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.material_defects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: material_defects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.material_defects_id_seq OWNED BY public.material_defects.id;


--
-- Name: material_locations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.material_locations (
    id bigint NOT NULL,
    project_id bigint,
    material_cost_id bigint,
    location_id bigint,
    location_type text,
    amount numeric
);


--
-- Name: material_locations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.material_locations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: material_locations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.material_locations_id_seq OWNED BY public.material_locations.id;


--
-- Name: materials; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.materials (
    id bigint NOT NULL,
    category text,
    code text,
    name text,
    unit text,
    notes text,
    has_serial_number boolean,
    article text,
    project_id bigint,
    planned_amount_for_project numeric,
    show_planned_amount_in_report boolean
);


--
-- Name: materials_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.materials_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: materials_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.materials_id_seq OWNED BY public.materials.id;


--
-- Name: mjd_objects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.mjd_objects (
    id bigint NOT NULL,
    model text,
    amount_stores bigint,
    amount_entrances bigint,
    has_basement boolean
);


--
-- Name: mjd_objects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.mjd_objects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: mjd_objects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.mjd_objects_id_seq OWNED BY public.mjd_objects.id;


--
-- Name: object_supervisors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.object_supervisors (
    id bigint NOT NULL,
    supervisor_worker_id bigint,
    object_id bigint
);


--
-- Name: object_supervisors_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.object_supervisors_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: object_supervisors_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.object_supervisors_id_seq OWNED BY public.object_supervisors.id;


--
-- Name: object_teams; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.object_teams (
    id bigint NOT NULL,
    team_id bigint,
    object_id bigint
);


--
-- Name: object_teams_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.object_teams_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: object_teams_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.object_teams_id_seq OWNED BY public.object_teams.id;


--
-- Name: objects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.objects (
    id bigint NOT NULL,
    object_detailed_id bigint,
    type text,
    name text,
    status text,
    project_id bigint
);


--
-- Name: objects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.objects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: objects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.objects_id_seq OWNED BY public.objects.id;


--
-- Name: operation_materials; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.operation_materials (
    id bigint NOT NULL,
    operation_id bigint,
    material_id bigint
);


--
-- Name: operation_materials_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.operation_materials_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: operation_materials_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.operation_materials_id_seq OWNED BY public.operation_materials.id;


--
-- Name: operations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.operations (
    id bigint NOT NULL,
    project_id bigint,
    name text,
    code text,
    cost_prime numeric(20,4),
    cost_with_customer numeric(20,4),
    planned_amount_for_project numeric,
    show_planned_amount_in_report boolean
);


--
-- Name: operations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.operations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: operations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.operations_id_seq OWNED BY public.operations.id;


--
-- Name: operator_error_founds; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.operator_error_founds (
    id bigint NOT NULL,
    invoice_materials_id bigint,
    material_cost_id bigint,
    amount numeric,
    notes text
);


--
-- Name: operator_error_founds_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.operator_error_founds_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: operator_error_founds_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.operator_error_founds_id_seq OWNED BY public.operator_error_founds.id;


--
-- Name: permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permissions (
    id bigint NOT NULL,
    role_id bigint,
    resource_id bigint,
    r boolean,
    w boolean,
    u boolean,
    d boolean
);


--
-- Name: permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.permissions_id_seq OWNED BY public.permissions.id;


--
-- Name: project_progress_materials; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_progress_materials (
    id bigint NOT NULL,
    project_id bigint,
    material_cost_id bigint,
    received numeric,
    installed numeric,
    amount_in_warehouse numeric,
    amount_in_teams numeric,
    amount_in_objects numeric,
    amount_write_off numeric,
    date timestamp with time zone
);


--
-- Name: project_progress_materials_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.project_progress_materials_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_progress_materials_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.project_progress_materials_id_seq OWNED BY public.project_progress_materials.id;


--
-- Name: project_progress_operations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_progress_operations (
    id bigint NOT NULL,
    project_id bigint,
    operation_id bigint,
    installed numeric,
    date timestamp with time zone
);


--
-- Name: project_progress_operations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.project_progress_operations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_progress_operations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.project_progress_operations_id_seq OWNED BY public.project_progress_operations.id;


--
-- Name: projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.projects (
    id bigint NOT NULL,
    name text,
    client text,
    budget numeric(20,2),
    budget_currency text,
    description text,
    signed_date_of_contract timestamp with time zone,
    date_start timestamp with time zone,
    date_end timestamp with time zone,
    project_manager text
);


--
-- Name: projects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.projects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: projects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.projects_id_seq OWNED BY public.projects.id;


--
-- Name: resources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.resources (
    id bigint NOT NULL,
    category text,
    name text,
    url text
);


--
-- Name: resources_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.resources_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: resources_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.resources_id_seq OWNED BY public.resources.id;


--
-- Name: roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.roles (
    id bigint NOT NULL,
    name text,
    description text
);


--
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.roles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.roles_id_seq OWNED BY public.roles.id;


--
-- Name: s_ip_objects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.s_ip_objects (
    id bigint NOT NULL,
    amount_feeders bigint
);


--
-- Name: s_ip_objects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.s_ip_objects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: s_ip_objects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.s_ip_objects_id_seq OWNED BY public.s_ip_objects.id;


--
-- Name: serial_number_locations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.serial_number_locations (
    id bigint NOT NULL,
    serial_number_id bigint,
    project_id bigint,
    location_id bigint,
    location_type text
);


--
-- Name: serial_number_locations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.serial_number_locations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: serial_number_locations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.serial_number_locations_id_seq OWNED BY public.serial_number_locations.id;


--
-- Name: serial_number_movements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.serial_number_movements (
    id bigint NOT NULL,
    serial_number_id bigint,
    project_id bigint,
    invoice_id bigint,
    invoice_type text,
    is_defected boolean,
    confirmation boolean
);


--
-- Name: serial_number_movements_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.serial_number_movements_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: serial_number_movements_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.serial_number_movements_id_seq OWNED BY public.serial_number_movements.id;


--
-- Name: serial_numbers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.serial_numbers (
    id bigint NOT NULL,
    project_id bigint,
    material_cost_id bigint,
    code text
);


--
-- Name: serial_numbers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.serial_numbers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: serial_numbers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.serial_numbers_id_seq OWNED BY public.serial_numbers.id;


--
-- Name: stvt_objects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stvt_objects (
    id bigint NOT NULL,
    voltage_class text,
    tt_coefficient text
);


--
-- Name: stvt_objects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.stvt_objects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: stvt_objects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.stvt_objects_id_seq OWNED BY public.stvt_objects.id;


--
-- Name: substation_cell_nourashes_substation_objects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.substation_cell_nourashes_substation_objects (
    id bigint NOT NULL,
    substation_object_id bigint,
    substation_cell_object_id bigint
);


--
-- Name: substation_cell_nourashes_substation_objects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.substation_cell_nourashes_substation_objects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: substation_cell_nourashes_substation_objects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.substation_cell_nourashes_substation_objects_id_seq OWNED BY public.substation_cell_nourashes_substation_objects.id;


--
-- Name: substation_cell_objects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.substation_cell_objects (
    id bigint NOT NULL
);


--
-- Name: substation_cell_objects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.substation_cell_objects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: substation_cell_objects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.substation_cell_objects_id_seq OWNED BY public.substation_cell_objects.id;


--
-- Name: substation_objects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.substation_objects (
    id bigint NOT NULL,
    voltage_class text,
    number_of_transformers bigint
);


--
-- Name: substation_objects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.substation_objects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: substation_objects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.substation_objects_id_seq OWNED BY public.substation_objects.id;


--
-- Name: team_leaders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.team_leaders (
    id bigint NOT NULL,
    team_id bigint,
    leader_worker_id bigint
);


--
-- Name: team_leaders_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.team_leaders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: team_leaders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.team_leaders_id_seq OWNED BY public.team_leaders.id;


--
-- Name: teams; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.teams (
    id bigint NOT NULL,
    project_id bigint,
    number text,
    mobile_number text,
    company text
);


--
-- Name: teams_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.teams_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: teams_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.teams_id_seq OWNED BY public.teams.id;


--
-- Name: tp_nourashes_objects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tp_nourashes_objects (
    id bigint NOT NULL,
    tp_object_id bigint,
    target_id bigint,
    target_type text
);


--
-- Name: tp_nourashes_objects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tp_nourashes_objects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tp_nourashes_objects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tp_nourashes_objects_id_seq OWNED BY public.tp_nourashes_objects.id;


--
-- Name: tp_objects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tp_objects (
    id bigint NOT NULL,
    model text,
    voltage_class text
);


--
-- Name: tp_objects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tp_objects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tp_objects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tp_objects_id_seq OWNED BY public.tp_objects.id;


--
-- Name: user_actions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_actions (
    id bigint NOT NULL,
    action_url text,
    action_type text,
    action_id bigint,
    action_status boolean,
    action_status_message text,
    user_id bigint,
    project_id bigint,
    date_of_action timestamp with time zone
);


--
-- Name: user_actions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_actions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_actions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_actions_id_seq OWNED BY public.user_actions.id;


--
-- Name: user_in_projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_in_projects (
    id bigint NOT NULL,
    project_id bigint,
    user_id bigint
);


--
-- Name: user_in_projects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_in_projects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_in_projects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_in_projects_id_seq OWNED BY public.user_in_projects.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    worker_id bigint,
    username text,
    password text,
    role_id bigint
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: worker_attendances; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.worker_attendances (
    id bigint NOT NULL,
    project_id bigint,
    worker_id bigint,
    start timestamp with time zone,
    "end" timestamp with time zone
);


--
-- Name: worker_attendances_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.worker_attendances_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: worker_attendances_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.worker_attendances_id_seq OWNED BY public.worker_attendances.id;


--
-- Name: workers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.workers (
    id bigint NOT NULL,
    project_id bigint,
    name text,
    company_worker_id text,
    job_title_in_company text,
    job_title_in_project text,
    mobile_number text
);


--
-- Name: workers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.workers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: workers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.workers_id_seq OWNED BY public.workers.id;


--
-- Name: auction_items id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auction_items ALTER COLUMN id SET DEFAULT nextval('public.auction_items_id_seq'::regclass);


--
-- Name: auction_packages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auction_packages ALTER COLUMN id SET DEFAULT nextval('public.auction_packages_id_seq'::regclass);


--
-- Name: auction_participant_prices id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auction_participant_prices ALTER COLUMN id SET DEFAULT nextval('public.auction_participant_prices_id_seq'::regclass);


--
-- Name: auctions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auctions ALTER COLUMN id SET DEFAULT nextval('public.auctions_id_seq'::regclass);


--
-- Name: districts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.districts ALTER COLUMN id SET DEFAULT nextval('public.districts_id_seq'::regclass);


--
-- Name: invoice_counts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_counts ALTER COLUMN id SET DEFAULT nextval('public.invoice_counts_id_seq'::regclass);


--
-- Name: invoice_inputs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_inputs ALTER COLUMN id SET DEFAULT nextval('public.invoice_inputs_id_seq'::regclass);


--
-- Name: invoice_materials id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_materials ALTER COLUMN id SET DEFAULT nextval('public.invoice_materials_id_seq'::regclass);


--
-- Name: invoice_object_operators id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_object_operators ALTER COLUMN id SET DEFAULT nextval('public.invoice_object_operators_id_seq'::regclass);


--
-- Name: invoice_objects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_objects ALTER COLUMN id SET DEFAULT nextval('public.invoice_objects_id_seq'::regclass);


--
-- Name: invoice_operations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_operations ALTER COLUMN id SET DEFAULT nextval('public.invoice_operations_id_seq'::regclass);


--
-- Name: invoice_output_out_of_projects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_output_out_of_projects ALTER COLUMN id SET DEFAULT nextval('public.invoice_output_out_of_projects_id_seq'::regclass);


--
-- Name: invoice_outputs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_outputs ALTER COLUMN id SET DEFAULT nextval('public.invoice_outputs_id_seq'::regclass);


--
-- Name: invoice_returns id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_returns ALTER COLUMN id SET DEFAULT nextval('public.invoice_returns_id_seq'::regclass);


--
-- Name: invoice_write_offs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_write_offs ALTER COLUMN id SET DEFAULT nextval('public.invoice_write_offs_id_seq'::regclass);


--
-- Name: kl04_kv_objects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.kl04_kv_objects ALTER COLUMN id SET DEFAULT nextval('public.kl04_kv_objects_id_seq'::regclass);


--
-- Name: material_costs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.material_costs ALTER COLUMN id SET DEFAULT nextval('public.material_costs_id_seq'::regclass);


--
-- Name: material_defects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.material_defects ALTER COLUMN id SET DEFAULT nextval('public.material_defects_id_seq'::regclass);


--
-- Name: material_locations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.material_locations ALTER COLUMN id SET DEFAULT nextval('public.material_locations_id_seq'::regclass);


--
-- Name: materials id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.materials ALTER COLUMN id SET DEFAULT nextval('public.materials_id_seq'::regclass);


--
-- Name: mjd_objects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mjd_objects ALTER COLUMN id SET DEFAULT nextval('public.mjd_objects_id_seq'::regclass);


--
-- Name: object_supervisors id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.object_supervisors ALTER COLUMN id SET DEFAULT nextval('public.object_supervisors_id_seq'::regclass);


--
-- Name: object_teams id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.object_teams ALTER COLUMN id SET DEFAULT nextval('public.object_teams_id_seq'::regclass);


--
-- Name: objects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.objects ALTER COLUMN id SET DEFAULT nextval('public.objects_id_seq'::regclass);


--
-- Name: operation_materials id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operation_materials ALTER COLUMN id SET DEFAULT nextval('public.operation_materials_id_seq'::regclass);


--
-- Name: operations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operations ALTER COLUMN id SET DEFAULT nextval('public.operations_id_seq'::regclass);


--
-- Name: operator_error_founds id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operator_error_founds ALTER COLUMN id SET DEFAULT nextval('public.operator_error_founds_id_seq'::regclass);


--
-- Name: permissions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions ALTER COLUMN id SET DEFAULT nextval('public.permissions_id_seq'::regclass);


--
-- Name: project_progress_materials id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_progress_materials ALTER COLUMN id SET DEFAULT nextval('public.project_progress_materials_id_seq'::regclass);


--
-- Name: project_progress_operations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_progress_operations ALTER COLUMN id SET DEFAULT nextval('public.project_progress_operations_id_seq'::regclass);


--
-- Name: projects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects ALTER COLUMN id SET DEFAULT nextval('public.projects_id_seq'::regclass);


--
-- Name: resources id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.resources ALTER COLUMN id SET DEFAULT nextval('public.resources_id_seq'::regclass);


--
-- Name: roles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles ALTER COLUMN id SET DEFAULT nextval('public.roles_id_seq'::regclass);


--
-- Name: s_ip_objects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.s_ip_objects ALTER COLUMN id SET DEFAULT nextval('public.s_ip_objects_id_seq'::regclass);


--
-- Name: serial_number_locations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.serial_number_locations ALTER COLUMN id SET DEFAULT nextval('public.serial_number_locations_id_seq'::regclass);


--
-- Name: serial_number_movements id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.serial_number_movements ALTER COLUMN id SET DEFAULT nextval('public.serial_number_movements_id_seq'::regclass);


--
-- Name: serial_numbers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.serial_numbers ALTER COLUMN id SET DEFAULT nextval('public.serial_numbers_id_seq'::regclass);


--
-- Name: stvt_objects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stvt_objects ALTER COLUMN id SET DEFAULT nextval('public.stvt_objects_id_seq'::regclass);


--
-- Name: substation_cell_nourashes_substation_objects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.substation_cell_nourashes_substation_objects ALTER COLUMN id SET DEFAULT nextval('public.substation_cell_nourashes_substation_objects_id_seq'::regclass);


--
-- Name: substation_cell_objects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.substation_cell_objects ALTER COLUMN id SET DEFAULT nextval('public.substation_cell_objects_id_seq'::regclass);


--
-- Name: substation_objects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.substation_objects ALTER COLUMN id SET DEFAULT nextval('public.substation_objects_id_seq'::regclass);


--
-- Name: team_leaders id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_leaders ALTER COLUMN id SET DEFAULT nextval('public.team_leaders_id_seq'::regclass);


--
-- Name: teams id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams ALTER COLUMN id SET DEFAULT nextval('public.teams_id_seq'::regclass);


--
-- Name: tp_nourashes_objects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tp_nourashes_objects ALTER COLUMN id SET DEFAULT nextval('public.tp_nourashes_objects_id_seq'::regclass);


--
-- Name: tp_objects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tp_objects ALTER COLUMN id SET DEFAULT nextval('public.tp_objects_id_seq'::regclass);


--
-- Name: user_actions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_actions ALTER COLUMN id SET DEFAULT nextval('public.user_actions_id_seq'::regclass);


--
-- Name: user_in_projects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_in_projects ALTER COLUMN id SET DEFAULT nextval('public.user_in_projects_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: worker_attendances id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.worker_attendances ALTER COLUMN id SET DEFAULT nextval('public.worker_attendances_id_seq'::regclass);


--
-- Name: workers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workers ALTER COLUMN id SET DEFAULT nextval('public.workers_id_seq'::regclass);


--
-- Name: auction_items auction_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auction_items
    ADD CONSTRAINT auction_items_pkey PRIMARY KEY (id);


--
-- Name: auction_packages auction_packages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auction_packages
    ADD CONSTRAINT auction_packages_pkey PRIMARY KEY (id);


--
-- Name: auction_participant_prices auction_participant_prices_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auction_participant_prices
    ADD CONSTRAINT auction_participant_prices_pkey PRIMARY KEY (id);


--
-- Name: auctions auctions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auctions
    ADD CONSTRAINT auctions_pkey PRIMARY KEY (id);


--
-- Name: districts districts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.districts
    ADD CONSTRAINT districts_pkey PRIMARY KEY (id);


--
-- Name: invoice_counts invoice_counts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_counts
    ADD CONSTRAINT invoice_counts_pkey PRIMARY KEY (id);


--
-- Name: invoice_inputs invoice_inputs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_inputs
    ADD CONSTRAINT invoice_inputs_pkey PRIMARY KEY (id);


--
-- Name: invoice_materials invoice_materials_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_materials
    ADD CONSTRAINT invoice_materials_pkey PRIMARY KEY (id);


--
-- Name: invoice_object_operators invoice_object_operators_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_object_operators
    ADD CONSTRAINT invoice_object_operators_pkey PRIMARY KEY (id);


--
-- Name: invoice_objects invoice_objects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_objects
    ADD CONSTRAINT invoice_objects_pkey PRIMARY KEY (id);


--
-- Name: invoice_operations invoice_operations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_operations
    ADD CONSTRAINT invoice_operations_pkey PRIMARY KEY (id);


--
-- Name: invoice_output_out_of_projects invoice_output_out_of_projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_output_out_of_projects
    ADD CONSTRAINT invoice_output_out_of_projects_pkey PRIMARY KEY (id);


--
-- Name: invoice_outputs invoice_outputs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_outputs
    ADD CONSTRAINT invoice_outputs_pkey PRIMARY KEY (id);


--
-- Name: invoice_returns invoice_returns_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_returns
    ADD CONSTRAINT invoice_returns_pkey PRIMARY KEY (id);


--
-- Name: invoice_write_offs invoice_write_offs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_write_offs
    ADD CONSTRAINT invoice_write_offs_pkey PRIMARY KEY (id);


--
-- Name: kl04_kv_objects kl04_kv_objects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.kl04_kv_objects
    ADD CONSTRAINT kl04_kv_objects_pkey PRIMARY KEY (id);


--
-- Name: material_costs material_costs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.material_costs
    ADD CONSTRAINT material_costs_pkey PRIMARY KEY (id);


--
-- Name: material_defects material_defects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.material_defects
    ADD CONSTRAINT material_defects_pkey PRIMARY KEY (id);


--
-- Name: material_locations material_locations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.material_locations
    ADD CONSTRAINT material_locations_pkey PRIMARY KEY (id);


--
-- Name: materials materials_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.materials
    ADD CONSTRAINT materials_pkey PRIMARY KEY (id);


--
-- Name: mjd_objects mjd_objects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mjd_objects
    ADD CONSTRAINT mjd_objects_pkey PRIMARY KEY (id);


--
-- Name: object_supervisors object_supervisors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.object_supervisors
    ADD CONSTRAINT object_supervisors_pkey PRIMARY KEY (id);


--
-- Name: object_teams object_teams_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.object_teams
    ADD CONSTRAINT object_teams_pkey PRIMARY KEY (id);


--
-- Name: objects objects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.objects
    ADD CONSTRAINT objects_pkey PRIMARY KEY (id);


--
-- Name: operation_materials operation_materials_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operation_materials
    ADD CONSTRAINT operation_materials_pkey PRIMARY KEY (id);


--
-- Name: operations operations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operations
    ADD CONSTRAINT operations_pkey PRIMARY KEY (id);


--
-- Name: operator_error_founds operator_error_founds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operator_error_founds
    ADD CONSTRAINT operator_error_founds_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: project_progress_materials project_progress_materials_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_progress_materials
    ADD CONSTRAINT project_progress_materials_pkey PRIMARY KEY (id);


--
-- Name: project_progress_operations project_progress_operations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_progress_operations
    ADD CONSTRAINT project_progress_operations_pkey PRIMARY KEY (id);


--
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);


--
-- Name: resources resources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.resources
    ADD CONSTRAINT resources_pkey PRIMARY KEY (id);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: s_ip_objects s_ip_objects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.s_ip_objects
    ADD CONSTRAINT s_ip_objects_pkey PRIMARY KEY (id);


--
-- Name: serial_number_locations serial_number_locations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.serial_number_locations
    ADD CONSTRAINT serial_number_locations_pkey PRIMARY KEY (id);


--
-- Name: serial_number_movements serial_number_movements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.serial_number_movements
    ADD CONSTRAINT serial_number_movements_pkey PRIMARY KEY (id);


--
-- Name: serial_numbers serial_numbers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.serial_numbers
    ADD CONSTRAINT serial_numbers_pkey PRIMARY KEY (id);


--
-- Name: stvt_objects stvt_objects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stvt_objects
    ADD CONSTRAINT stvt_objects_pkey PRIMARY KEY (id);


--
-- Name: substation_cell_nourashes_substation_objects substation_cell_nourashes_substation_objects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.substation_cell_nourashes_substation_objects
    ADD CONSTRAINT substation_cell_nourashes_substation_objects_pkey PRIMARY KEY (id);


--
-- Name: substation_cell_objects substation_cell_objects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.substation_cell_objects
    ADD CONSTRAINT substation_cell_objects_pkey PRIMARY KEY (id);


--
-- Name: substation_objects substation_objects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.substation_objects
    ADD CONSTRAINT substation_objects_pkey PRIMARY KEY (id);


--
-- Name: team_leaders team_leaders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_leaders
    ADD CONSTRAINT team_leaders_pkey PRIMARY KEY (id);


--
-- Name: teams teams_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_pkey PRIMARY KEY (id);


--
-- Name: tp_nourashes_objects tp_nourashes_objects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tp_nourashes_objects
    ADD CONSTRAINT tp_nourashes_objects_pkey PRIMARY KEY (id);


--
-- Name: tp_objects tp_objects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tp_objects
    ADD CONSTRAINT tp_objects_pkey PRIMARY KEY (id);


--
-- Name: user_actions user_actions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_actions
    ADD CONSTRAINT user_actions_pkey PRIMARY KEY (id);


--
-- Name: user_in_projects user_in_projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_in_projects
    ADD CONSTRAINT user_in_projects_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: worker_attendances worker_attendances_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.worker_attendances
    ADD CONSTRAINT worker_attendances_pkey PRIMARY KEY (id);


--
-- Name: workers workers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workers
    ADD CONSTRAINT workers_pkey PRIMARY KEY (id);


--
-- Name: auction_items fk_auction_packages_auction_items; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auction_items
    ADD CONSTRAINT fk_auction_packages_auction_items FOREIGN KEY (auction_package_id) REFERENCES public.auction_packages(id);


--
-- Name: auction_packages fk_auctions_auction_packages; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auction_packages
    ADD CONSTRAINT fk_auctions_auction_packages FOREIGN KEY (auction_id) REFERENCES public.auctions(id);


--
-- Name: invoice_objects fk_districts_invoice_object; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_objects
    ADD CONSTRAINT fk_districts_invoice_object FOREIGN KEY (district_id) REFERENCES public.districts(id);


--
-- Name: invoice_outputs fk_districts_invoice_output; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_outputs
    ADD CONSTRAINT fk_districts_invoice_output FOREIGN KEY (district_id) REFERENCES public.districts(id);


--
-- Name: invoice_object_operators fk_invoice_objects_invoice_object_operators; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_object_operators
    ADD CONSTRAINT fk_invoice_objects_invoice_object_operators FOREIGN KEY (invoice_object_id) REFERENCES public.invoice_objects(id);


--
-- Name: invoice_materials fk_material_costs_invoice_materials; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_materials
    ADD CONSTRAINT fk_material_costs_invoice_materials FOREIGN KEY (material_cost_id) REFERENCES public.material_costs(id);


--
-- Name: material_locations fk_material_costs_material_locations; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.material_locations
    ADD CONSTRAINT fk_material_costs_material_locations FOREIGN KEY (material_cost_id) REFERENCES public.material_costs(id);


--
-- Name: project_progress_materials fk_material_costs_project_progresses_materials; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_progress_materials
    ADD CONSTRAINT fk_material_costs_project_progresses_materials FOREIGN KEY (material_cost_id) REFERENCES public.material_costs(id);


--
-- Name: material_defects fk_material_locations_material_defects; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.material_defects
    ADD CONSTRAINT fk_material_locations_material_defects FOREIGN KEY (material_location_id) REFERENCES public.material_locations(id);


--
-- Name: material_costs fk_materials_material_costs; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.material_costs
    ADD CONSTRAINT fk_materials_material_costs FOREIGN KEY (material_id) REFERENCES public.materials(id);


--
-- Name: operation_materials fk_materials_operation_materials; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operation_materials
    ADD CONSTRAINT fk_materials_operation_materials FOREIGN KEY (material_id) REFERENCES public.materials(id);


--
-- Name: invoice_objects fk_objects_invoice_object; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_objects
    ADD CONSTRAINT fk_objects_invoice_object FOREIGN KEY (object_id) REFERENCES public.objects(id);


--
-- Name: object_supervisors fk_objects_object_supervisors; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.object_supervisors
    ADD CONSTRAINT fk_objects_object_supervisors FOREIGN KEY (object_id) REFERENCES public.objects(id);


--
-- Name: object_teams fk_objects_object_teams; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.object_teams
    ADD CONSTRAINT fk_objects_object_teams FOREIGN KEY (object_id) REFERENCES public.objects(id);


--
-- Name: substation_cell_nourashes_substation_objects fk_objects_substation_cell_objects; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.substation_cell_nourashes_substation_objects
    ADD CONSTRAINT fk_objects_substation_cell_objects FOREIGN KEY (substation_cell_object_id) REFERENCES public.objects(id);


--
-- Name: substation_cell_nourashes_substation_objects fk_objects_substation_objects; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.substation_cell_nourashes_substation_objects
    ADD CONSTRAINT fk_objects_substation_objects FOREIGN KEY (substation_object_id) REFERENCES public.objects(id);


--
-- Name: tp_nourashes_objects fk_objects_tp_nourashes_objects; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tp_nourashes_objects
    ADD CONSTRAINT fk_objects_tp_nourashes_objects FOREIGN KEY (tp_object_id) REFERENCES public.objects(id);


--
-- Name: invoice_operations fk_operations_invoice_operations; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_operations
    ADD CONSTRAINT fk_operations_invoice_operations FOREIGN KEY (operation_id) REFERENCES public.operations(id);


--
-- Name: operation_materials fk_operations_operation_materials; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operation_materials
    ADD CONSTRAINT fk_operations_operation_materials FOREIGN KEY (operation_id) REFERENCES public.operations(id);


--
-- Name: project_progress_operations fk_operations_project_progress_operations; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_progress_operations
    ADD CONSTRAINT fk_operations_project_progress_operations FOREIGN KEY (operation_id) REFERENCES public.operations(id);


--
-- Name: districts fk_projects_districts; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.districts
    ADD CONSTRAINT fk_projects_districts FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: invoice_counts fk_projects_invoice_counts; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_counts
    ADD CONSTRAINT fk_projects_invoice_counts FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: invoice_inputs fk_projects_invoice_inputs; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_inputs
    ADD CONSTRAINT fk_projects_invoice_inputs FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: invoice_materials fk_projects_invoice_materials; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_materials
    ADD CONSTRAINT fk_projects_invoice_materials FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: invoice_objects fk_projects_invoice_object; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_objects
    ADD CONSTRAINT fk_projects_invoice_object FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: invoice_operations fk_projects_invoice_operations; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_operations
    ADD CONSTRAINT fk_projects_invoice_operations FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: invoice_output_out_of_projects fk_projects_invoice_output_out_of_projects; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_output_out_of_projects
    ADD CONSTRAINT fk_projects_invoice_output_out_of_projects FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: invoice_outputs fk_projects_invoice_outputs; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_outputs
    ADD CONSTRAINT fk_projects_invoice_outputs FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: invoice_returns fk_projects_invoice_returns; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_returns
    ADD CONSTRAINT fk_projects_invoice_returns FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: invoice_write_offs fk_projects_invoice_write_offs; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_write_offs
    ADD CONSTRAINT fk_projects_invoice_write_offs FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: material_locations fk_projects_material_locations; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.material_locations
    ADD CONSTRAINT fk_projects_material_locations FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: materials fk_projects_materials; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.materials
    ADD CONSTRAINT fk_projects_materials FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: objects fk_projects_objects; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.objects
    ADD CONSTRAINT fk_projects_objects FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: operations fk_projects_operations; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operations
    ADD CONSTRAINT fk_projects_operations FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: project_progress_materials fk_projects_project_progress_materials; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_progress_materials
    ADD CONSTRAINT fk_projects_project_progress_materials FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: project_progress_operations fk_projects_project_progress_operations; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_progress_operations
    ADD CONSTRAINT fk_projects_project_progress_operations FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: serial_number_locations fk_projects_serial_number_locations; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.serial_number_locations
    ADD CONSTRAINT fk_projects_serial_number_locations FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: serial_number_movements fk_projects_serial_number_movements; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.serial_number_movements
    ADD CONSTRAINT fk_projects_serial_number_movements FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: serial_numbers fk_projects_serial_numbers; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.serial_numbers
    ADD CONSTRAINT fk_projects_serial_numbers FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: teams fk_projects_teams; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT fk_projects_teams FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: user_actions fk_projects_user_actions; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_actions
    ADD CONSTRAINT fk_projects_user_actions FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: user_in_projects fk_projects_user_in_projects; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_in_projects
    ADD CONSTRAINT fk_projects_user_in_projects FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: worker_attendances fk_projects_worker_attendances; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.worker_attendances
    ADD CONSTRAINT fk_projects_worker_attendances FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: permissions fk_resources_permissions; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT fk_resources_permissions FOREIGN KEY (resource_id) REFERENCES public.resources(id);


--
-- Name: permissions fk_roles_permissions; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT fk_roles_permissions FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: users fk_roles_users; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT fk_roles_users FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: serial_number_locations fk_serial_numbers_serial_number_locations; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.serial_number_locations
    ADD CONSTRAINT fk_serial_numbers_serial_number_locations FOREIGN KEY (serial_number_id) REFERENCES public.serial_numbers(id);


--
-- Name: serial_number_movements fk_serial_numbers_serial_number_movements; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.serial_number_movements
    ADD CONSTRAINT fk_serial_numbers_serial_number_movements FOREIGN KEY (serial_number_id) REFERENCES public.serial_numbers(id);


--
-- Name: invoice_objects fk_teams_invoice_object; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_objects
    ADD CONSTRAINT fk_teams_invoice_object FOREIGN KEY (team_id) REFERENCES public.teams(id);


--
-- Name: invoice_outputs fk_teams_invoice_outputs; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_outputs
    ADD CONSTRAINT fk_teams_invoice_outputs FOREIGN KEY (team_id) REFERENCES public.teams(id);


--
-- Name: object_teams fk_teams_object_teams; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.object_teams
    ADD CONSTRAINT fk_teams_object_teams FOREIGN KEY (team_id) REFERENCES public.teams(id);


--
-- Name: team_leaders fk_teams_team_leaderss; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_leaders
    ADD CONSTRAINT fk_teams_team_leaderss FOREIGN KEY (team_id) REFERENCES public.teams(id);


--
-- Name: auction_participant_prices fk_users_auction_participant_prices; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auction_participant_prices
    ADD CONSTRAINT fk_users_auction_participant_prices FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: user_actions fk_users_user_actions; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_actions
    ADD CONSTRAINT fk_users_user_actions FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: user_in_projects fk_users_user_in_projects; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_in_projects
    ADD CONSTRAINT fk_users_user_in_projects FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: invoice_inputs fk_workers_invoice_inputs_released; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_inputs
    ADD CONSTRAINT fk_workers_invoice_inputs_released FOREIGN KEY (released_worker_id) REFERENCES public.workers(id);


--
-- Name: invoice_inputs fk_workers_invoice_inputs_warehouse_manager; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_inputs
    ADD CONSTRAINT fk_workers_invoice_inputs_warehouse_manager FOREIGN KEY (warehouse_manager_worker_id) REFERENCES public.workers(id);


--
-- Name: invoice_object_operators fk_workers_invoice_object_operators; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_object_operators
    ADD CONSTRAINT fk_workers_invoice_object_operators FOREIGN KEY (operator_worker_id) REFERENCES public.workers(id);


--
-- Name: invoice_objects fk_workers_invoice_objects_supervisor; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_objects
    ADD CONSTRAINT fk_workers_invoice_objects_supervisor FOREIGN KEY (supervisor_worker_id) REFERENCES public.workers(id);


--
-- Name: invoice_output_out_of_projects fk_workers_invoice_output_out_of_project_released; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_output_out_of_projects
    ADD CONSTRAINT fk_workers_invoice_output_out_of_project_released FOREIGN KEY (released_worker_id) REFERENCES public.workers(id);


--
-- Name: invoice_outputs fk_workers_invoice_outputs_recipient; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_outputs
    ADD CONSTRAINT fk_workers_invoice_outputs_recipient FOREIGN KEY (recipient_worker_id) REFERENCES public.workers(id);


--
-- Name: invoice_outputs fk_workers_invoice_outputs_released; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_outputs
    ADD CONSTRAINT fk_workers_invoice_outputs_released FOREIGN KEY (released_worker_id) REFERENCES public.workers(id);


--
-- Name: invoice_outputs fk_workers_invoice_outputs_warehouse_manager; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_outputs
    ADD CONSTRAINT fk_workers_invoice_outputs_warehouse_manager FOREIGN KEY (warehouse_manager_worker_id) REFERENCES public.workers(id);


--
-- Name: invoice_returns fk_workers_invoice_returns; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_returns
    ADD CONSTRAINT fk_workers_invoice_returns FOREIGN KEY (accepted_by_worker_id) REFERENCES public.workers(id);


--
-- Name: invoice_write_offs fk_workers_invoice_write_off_releaseds; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_write_offs
    ADD CONSTRAINT fk_workers_invoice_write_off_releaseds FOREIGN KEY (released_worker_id) REFERENCES public.workers(id);


--
-- Name: object_supervisors fk_workers_object_supervisors; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.object_supervisors
    ADD CONSTRAINT fk_workers_object_supervisors FOREIGN KEY (supervisor_worker_id) REFERENCES public.workers(id);


--
-- Name: team_leaders fk_workers_team_leaderss; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_leaders
    ADD CONSTRAINT fk_workers_team_leaderss FOREIGN KEY (leader_worker_id) REFERENCES public.workers(id);


--
-- Name: users fk_workers_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT fk_workers_user FOREIGN KEY (worker_id) REFERENCES public.workers(id);


--
-- Name: worker_attendances fk_workers_worker_attendances; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.worker_attendances
    ADD CONSTRAINT fk_workers_worker_attendances FOREIGN KEY (worker_id) REFERENCES public.workers(id);


--
-- PostgreSQL database dump complete
--



-- +goose Down
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
