# Use the official PostgreSQL image as the base
FROM postgres:latest

# Set environment variables
ENV POSTGRES_USER=myuser
ENV POSTGRES_PASSWORD=password
ENV POSTGRES_DB=postgres

# Expose the PostgreSQL port (default is 5432)
# EXPOSE 5432
EXPOSE 5433

# You can add custom PostgreSQL configuration here if needed
# For example, to change the listening address:
# COPY postgresql.conf /etc/postgresql/postgresql.conf
#
# Or to run initialization scripts:
# COPY init.sql /docker-entrypoint-initdb.d/

# The official PostgreSQL image handles starting the database when the container runs
# No explicit CMD instruction is needed here for basic usage.