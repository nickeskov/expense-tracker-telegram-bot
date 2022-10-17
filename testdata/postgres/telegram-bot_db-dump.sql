--
-- PostgreSQL database dump
--

-- Dumped from database version 15.0
-- Dumped by pg_dump version 15.0

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

--
-- Data for Name: exchange_rates; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.exchange_rates (id, currency, date, rate) FROM stdin;
25	EUR	2022-10-06	0.01687400
26	RUB	2022-10-06	1.00000000
27	USD	2022-10-06	0.01652500
1	RUB	2022-10-17	1.00000000
3	USD	2022-10-17	0.01600400
2	EUR	2022-10-17	0.01641900
157	USD	2022-10-07	0.01604800
158	RUB	2022-10-07	1.00000000
159	EUR	2022-10-07	0.01647900
160	RUB	2022-10-08	1.00000000
161	EUR	2022-10-08	0.01647000
162	USD	2022-10-08	0.01606200
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, currency, monthly_limit) FROM stdin;
204257863	RUB	\N
\.


--
-- Data for Name: expenses; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.expenses (id, user_id, category, amount, date, comment) FROM stdin;
1	204257863	test-category	100.00000	2022-10-06	test-comment
2	204257863	test-category	100.00000	2022-10-06	test-comment
3	204257863	test-category	100.00000	2022-10-06	test-comment
4	204257863	test-category	100.00000	2022-10-06	test-comment
5	204257863	test-category	100.00000	2022-10-06	test-comment
6	204257863	test-category	100.00000	2022-10-06	test-comment
7	204257863	test-category	100.00000	2022-10-06	test-comment
9	204257863	test-category	605.14372	2022-10-06	test-comment
10	204257863	test-category	605.14372	2022-10-06	test-comment
11	204257863	test-category	100000000000000.00000	2022-10-07	test-comment
12	204257863	test-category	100000000000000.00000	2022-10-07	test-comment
13	204257863	test-category	100.10000	2022-10-08	test-comment
14	204257863	test-test-category	100.10000	2022-10-08	test-comment
15	204257863	test-test-category	234.00000	2022-10-08	test-comment
16	204257863	test-test-category	234.00000	2022-10-12	test-comment
\.


--
-- Data for Name: goose_db_version; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.goose_db_version (id, version_id, is_applied, tstamp) FROM stdin;
1	0	t	2022-10-17 02:27:58.608892
6	20221016225816	t	2022-10-17 04:03:30.741566
\.


--
-- Name: exchange_rates_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.exchange_rates_id_seq', 318, true);


--
-- Name: expenses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.expenses_id_seq', 16, true);


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.goose_db_version_id_seq', 6, true);


--
-- PostgreSQL database dump complete
--

