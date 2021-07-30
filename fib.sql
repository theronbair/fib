CREATE DATABASE fib WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'C' LC_CTYPE = 'C';

\connect fib

SELECT pg_catalog.set_config('search_path', 'public', false);

CREATE TABLE public.fib_memo (
    n bigint unique,
    val bigint
);

CREATE USER fib with password 'fib';
GRANT ALL PRIVILEGES ON fib_memo to fib;
ALTER TABLE public.fib_memo OWNER TO fib;

