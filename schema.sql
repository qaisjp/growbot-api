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

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: rooms; Type: TABLE; Schema: public; Owner: growbot
--

CREATE TABLE public.rooms (
    id integer NOT NULL,
    user_id integer NOT NULL,
    name text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at time without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.rooms OWNER TO growbot;

--
-- Name: rooms_id_seq; Type: SEQUENCE; Schema: public; Owner: growbot
--

CREATE SEQUENCE public.rooms_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.rooms_id_seq OWNER TO growbot;

--
-- Name: rooms_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: growbot
--

ALTER SEQUENCE public.rooms_id_seq OWNED BY public.rooms.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: growbot
--

CREATE TABLE public.users (
    id integer NOT NULL,
    forename text NOT NULL,
    surname text NOT NULL,
    password character(60) NOT NULL,
    email character varying(254) NOT NULL,
    is_verified boolean DEFAULT false NOT NULL,
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
-- Name: rooms id; Type: DEFAULT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.rooms ALTER COLUMN id SET DEFAULT nextval('public.rooms_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: rooms rooms_id_pkey; Type: CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.rooms
    ADD CONSTRAINT rooms_id_pkey PRIMARY KEY (id);


--
-- Name: rooms rooms_user_id_name_key; Type: CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.rooms
    ADD CONSTRAINT rooms_user_id_name_key UNIQUE (user_id, name);


--
-- Name: CONSTRAINT rooms_user_id_name_key ON rooms; Type: COMMENT; Schema: public; Owner: growbot
--

COMMENT ON CONSTRAINT rooms_user_id_name_key ON public.rooms IS 'A user cannot have multiple rooms with the same name';


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
-- Name: rooms rooms_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: growbot
--

ALTER TABLE ONLY public.rooms
    ADD CONSTRAINT rooms_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

