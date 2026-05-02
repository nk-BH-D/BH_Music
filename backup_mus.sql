--
-- PostgreSQL database dump
--

\restrict j2uwhenglT5QlmC03oa0IE2mZmXBkBgZUoSWCLKZOTm4UdvX22o5eUdD89wKcRa

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
-- Name: tracks; Type: TABLE; Schema: public; Owner: old_man
--

CREATE TABLE public.tracks (
    file_unique_id text NOT NULL,
    file_id text NOT NULL,
    title text NOT NULL,
    artist text NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.tracks OWNER TO old_man;

--
-- Data for Name: tracks; Type: TABLE DATA; Schema: public; Owner: old_man
--

COPY public.tracks (file_unique_id, file_id, title, artist, created_at) FROM stdin;
AgADVJQAAkR_KUs	CQACAgIAAxkBAAMPaeT84rYbYXiYGf-DBJVdPDWFLAQAAlSUAAJEfylLB9HlR-v8ENw7BA	Jealous	9mice, ЕГОР КРИД, тёмный принц, madk1d	2026-04-19 16:03:43.864205+00
AgAD6ZgAAkXSSUs	CQACAgIAAxkDAAM4aekxqnJr9fPhV1cHXXVzRWKLE4MAAumYAAJF0klL3cIs5ofI9zA7BA	Беспечный Ангел	Ария	2026-04-22 20:37:59.173779+00
AgADu5EAAh_kkEs	CQACAgIAAxkDAANFafEUbZe1tKTMJCaOX8IrUIgQPBoAAruRAAIf5JBLLv-4JdhdOEM7BA	8 Миля	Madk1D VILLIAN	2026-04-28 20:11:21.410563+00
AgADwJEAAh_kkEs	CQACAgIAAxkDAANJafEUzUuGDtQm5yXs4W19ZNCexzYAAsCRAAIf5JBLZeYGgLB7_807BA	MARTINE ROSE	Madk1D Greyrock Tewiq	2026-04-28 20:12:56.898015+00
\.


--
-- Name: tracks tracks_pkey; Type: CONSTRAINT; Schema: public; Owner: old_man
--

ALTER TABLE ONLY public.tracks
    ADD CONSTRAINT tracks_pkey PRIMARY KEY (file_unique_id);


--
-- Name: tracks tracks_title_artist_key; Type: CONSTRAINT; Schema: public; Owner: old_man
--

ALTER TABLE ONLY public.tracks
    ADD CONSTRAINT tracks_title_artist_key UNIQUE (title, artist);


--
-- Name: uniq_track; Type: INDEX; Schema: public; Owner: old_man
--

CREATE UNIQUE INDEX uniq_track ON public.tracks USING btree (lower(title), lower(artist));


--
-- PostgreSQL database dump complete
--

\unrestrict j2uwhenglT5QlmC03oa0IE2mZmXBkBgZUoSWCLKZOTm4UdvX22o5eUdD89wKcRa

