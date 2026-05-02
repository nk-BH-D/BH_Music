--
-- PostgreSQL database dump
--

\restrict 2qrZUmmTvs6UfFV8E9l6kz539zgdDYKV9pzucTpdcoF2pZlSgd480XomwXy9K9Z

-- Dumped from database version 15.14
-- Dumped by pg_dump version 15.14

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: users; Type: TABLE; Schema: public; Owner: user_peeper
--

CREATE TABLE public.users (
    user_id bigint NOT NULL,
    chat_id bigint,
    full_name text NOT NULL,
    login text,
    status text NOT NULL,
    file_id_list text[] DEFAULT '{}'::text[] NOT NULL,
    created_ad timestamp with time zone DEFAULT now()
);


ALTER TABLE public.users OWNER TO user_peeper;

--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: user_peeper
--

COPY public.users (user_id, chat_id, full_name, login, status, file_id_list, created_ad) FROM stdin;
2009375704	2009375704	♱Burning_Head♱	nk_BH_D	nuser	{CQACAgIAAxkBAAMPaeT84rYbYXiYGf-DBJVdPDWFLAQAAlSUAAJEfylLB9HlR-v8ENw7BA,CQACAgIAAxkDAAM4aekxqnJr9fPhV1cHXXVzRWKLE4MAAumYAAJF0klL3cIs5ofI9zA7BA,CQACAgIAAxkDAANFafEUbZe1tKTMJCaOX8IrUIgQPBoAAruRAAIf5JBLLv-4JdhdOEM7BA,CQACAgIAAxkDAANJafEUzUuGDtQm5yXs4W19ZNCexzYAAsCRAAIf5JBLZeYGgLB7_807BA}	2026-04-19 15:27:48.891392+00
8429485588	8429485588	oldman	o_o_nan_o_o	nuser	{}	2026-04-28 20:16:17.951045+00
\.


--
-- Name: users users_login_key; Type: CONSTRAINT; Schema: public; Owner: user_peeper
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_login_key UNIQUE (login);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: user_peeper
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- PostgreSQL database dump complete
--

\unrestrict 2qrZUmmTvs6UfFV8E9l6kz539zgdDYKV9pzucTpdcoF2pZlSgd480XomwXy9K9Z

