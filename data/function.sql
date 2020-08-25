--
-- Name: convto360(double precision); Type: FUNCTION; Schema: public; Owner: postgres
--
DROP FUNCTION if exists  public.convto360(x double precision);
CREATE FUNCTION public.convto360(x double precision) RETURNS double precision
    LANGUAGE plpgsql
AS $$
begin
    if x < 0 then
        return x + 360;
    else
        return x;
    end if;
end
$$;


ALTER FUNCTION public.convto360(x double precision) OWNER TO postgres;

--
-- Name: accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

--DROP SEQUENCE if exists public.accounts_id_seq;
--CREATE SEQUENCE public.accounts_id_seq
--    AS integer
--    START WITH 1
--    INCREMENT BY 1
--    NO MINVALUE
--    NO MAXVALUE
--    CACHE 1;


--ALTER TABLE public.accounts_id_seq OWNER TO postgres;

--
-- Name: accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

-- ALTER SEQUENCE public.accounts_id_seq OWNED BY public.accounts.id;
