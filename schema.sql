--
-- PostgreSQL database dump
--

-- Dumped from database version 11.1
-- Dumped by pg_dump version 11.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: event_action_name; Type: TYPE; Schema: public; Owner: growbot
--

CREATE TYPE public.event_action_name AS ENUM (
    'PLANT_WATER',
    'PLANT_CAPTURE_PHOTO',
    'ROBOT_RANDOM_MOVEMENT'
);


ALTER TYPE public.event_action_name OWNER TO growbot;

--
-- Name: growbot_create_state(); Type: FUNCTION; Schema: public; Owner: growbot
--

CREATE FUNCTION public.growbot_create_state() RETURNS trigger
    LANGUAGE plpgsql
    AS $$begin
	insert into public.robot_state(id) values (new.id);
	return new;
end;$$;


ALTER FUNCTION public.growbot_create_state() OWNER TO growbot;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: event_actions; Type: TABLE; Schema: public; Owner: growbot
--

CREATE TABLE public.event_actions (
    id integer NOT NULL,
    name public.event_action_name NOT NULL,
    data jsonb NOT NULL,
    plant_id uuid,
    event_id integer NOT NULL,
    robot_id uuid NOT NULL
);


ALTER TABLE public.event_actions OWNER TO growbot;

--
-- Name: event_actions_id_seq; Type: SEQUENCE; Schema: public; Owner: growbot
--

CREATE SEQUENCE public.event_actions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.event_actions_id_seq OWNER TO growbot;

--
-- Name: event_actions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: growbot
--

ALTER SEQUENCE public.event_actions_id_seq OWNED BY public.event_actions.id;


--
-- Name: events; Type: TABLE; Schema: public; Owner: growbot
--

CREATE TABLE public.events (
    id integer NOT NULL,
    summary text NOT NULL,
    recurrence text[] DEFAULT ARRAY[]::text[] NOT NULL,
    user_id integer NOT NULL
);


ALTER TABLE public.events OWNER TO growbot;

--
-- Name: events_id_seq; Type: SEQUENCE; Schema: public; Owner: growbot
--

CREATE SEQUENCE public.events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.events_id_seq OWNER TO growbot;

--
-- Name: events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: growbot
--

ALTER SEQUENCE public.events_id_seq OWNED BY public.events.id;


--
-- Name: log; Type: TABLE; Schema: public; Owner: growbot
--

CREATE TABLE public.log (
    id integer NOT NULL,
    user_id integer NOT NULL,
    message text NOT NULL,
    severity integer NOT NULL,
    robot_id uuid,
    plant_id uuid,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.log OWNER TO growbot;

--
-- Name: log_id_seq; Type: SEQUENCE; Schema: public; Owner: growbot
--

CREATE SEQUENCE public.log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.log_id_seq OWNER TO growbot;

--
-- Name: log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: growbot
--

ALTER SEQUENCE public.log_id_seq OWNED BY public.log.id;


--
-- Name: plant_photos; Type: TABLE; Schema: public; Owner: growbot
--

CREATE TABLE public.plant_photos (
    id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    filename uuid NOT NULL,
    plant_id uuid NOT NULL
);


ALTER TABLE public.plant_photos OWNER TO growbot;

--
-- Name: plant_photos_id_seq; Type: SEQUENCE; Schema: public; Owner: growbot
--

CREATE SEQUENCE public.plant_photos_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.plant_photos_id_seq OWNER TO growbot;

--
-- Name: plant_photos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: growbot
--

ALTER SEQUENCE public.plant_photos_id_seq OWNED BY public.plant_photos.id;


--
-- Name: plants; Type: TABLE; Schema: public; Owner: growbot
--

CREATE TABLE public.plants (
    id uuid NOT NULL,
    user_id integer NOT NULL,
    name text NOT NULL
);


ALTER TABLE public.plants OWNER TO growbot;

--
-- Name: robot_state; Type: TABLE; Schema: public; Owner: growbot
--

CREATE TABLE public.robot_state (
    id uuid NOT NULL,
    battery_level integer DEFAULT 0 NOT NULL,
    water_level integer DEFAULT 0 NOT NULL,
    distress boolean DEFAULT false NOT NULL,
    seen_at timestamp without time zone
);


ALTER TABLE public.robot_state OWNER TO growbot;

--
-- Name: TABLE robot_state; Type: COMMENT; Schema: public; Owner: growbot
--

COMMENT ON TABLE public.robot_state IS 'All fields are mandatory, except that seen_at has the special case where if it''s `null`, the robot has not been "seen" yet. That is to say, the robot hasn''t been turned on yet.

This is so that the interactive interfaces can report "Not seen yet" instead of just the default (blank) values.';


--
-- Name: robots; Type: TABLE; Schema: public; Owner: growbot
--

CREATE TABLE public.robots (
    id uuid NOT NULL,
    admin_token text NOT NULL,
    user_id integer,
    title text DEFAULT ''::text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.robots OWNER TO growbot;

--
-- Name: users; Type: TABLE; Schema: public; Owner: growbot
--

CREATE TABLE public.users (
    id integer NOT NULL,
    forename text NOT NULL,
    surname text NOT NULL,
    password character(60) NOT NULL,
    email character varying(254) NOT NULL,
    is_activated boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.users OWNER TO growbot;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: growbot
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_id_seq OWNER TO growbot;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: growbot
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: event_actions id; Type: DEFAULT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.event_actions ALTER COLUMN id SET DEFAULT nextval('public.event_actions_id_seq'::regclass);


--
-- Name: events id; Type: DEFAULT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.events ALTER COLUMN id SET DEFAULT nextval('public.events_id_seq'::regclass);


--
-- Name: log id; Type: DEFAULT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.log ALTER COLUMN id SET DEFAULT nextval('public.log_id_seq'::regclass);


--
-- Name: plant_photos id; Type: DEFAULT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.plant_photos ALTER COLUMN id SET DEFAULT nextval('public.plant_photos_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: event_actions event_actions_id_key; Type: CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.event_actions
    ADD CONSTRAINT event_actions_id_key PRIMARY KEY (id);


--
-- Name: events events_id_key; Type: CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_id_key PRIMARY KEY (id);


--
-- Name: log log_id_pkey; Type: CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.log
    ADD CONSTRAINT log_id_pkey PRIMARY KEY (id);


--
-- Name: plant_photos plant_photos_id_key; Type: CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.plant_photos
    ADD CONSTRAINT plant_photos_id_key PRIMARY KEY (id);


--
-- Name: plants plants_id_pkey; Type: CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.plants
    ADD CONSTRAINT plants_id_pkey PRIMARY KEY (id);


--
-- Name: plants plants_name_user_id_key; Type: CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.plants
    ADD CONSTRAINT plants_name_user_id_key UNIQUE (user_id, name);


--
-- Name: robot_state robot_state_id_pkey; Type: CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.robot_state
    ADD CONSTRAINT robot_state_id_pkey PRIMARY KEY (id);


--
-- Name: robots robots_id_pkey; Type: CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.robots
    ADD CONSTRAINT robots_id_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_id_pkey; Type: CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_id_pkey PRIMARY KEY (id);


--
-- Name: robots trig_create_state; Type: TRIGGER; Schema: public; Owner: growbot
--

CREATE TRIGGER trig_create_state AFTER INSERT ON public.robots FOR EACH ROW EXECUTE PROCEDURE public.growbot_create_state();


--
-- Name: event_actions event_actions_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.event_actions
    ADD CONSTRAINT event_actions_event_id_fkey FOREIGN KEY (event_id) REFERENCES public.events(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: event_actions event_actions_plant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.event_actions
    ADD CONSTRAINT event_actions_plant_id_fkey FOREIGN KEY (plant_id) REFERENCES public.plants(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: event_actions event_actions_robot_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.event_actions
    ADD CONSTRAINT event_actions_robot_id_fkey FOREIGN KEY (robot_id) REFERENCES public.robots(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: events events_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: log log_plant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.log
    ADD CONSTRAINT log_plant_id_fkey FOREIGN KEY (plant_id) REFERENCES public.plants(id);


--
-- Name: log log_robot_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.log
    ADD CONSTRAINT log_robot_id_fkey FOREIGN KEY (robot_id) REFERENCES public.robots(id);


--
-- Name: log log_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.log
    ADD CONSTRAINT log_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: plant_photos plant_photos_plant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.plant_photos
    ADD CONSTRAINT plant_photos_plant_id_fkey FOREIGN KEY (plant_id) REFERENCES public.plants(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: plants plants_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.plants
    ADD CONSTRAINT plants_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: robot_state robot_state_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.robot_state
    ADD CONSTRAINT robot_state_id_fkey FOREIGN KEY (id) REFERENCES public.robots(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: robots robots_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.robots
    ADD CONSTRAINT robots_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

