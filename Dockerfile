version: "3.9"

services:
  app:
    build: .
    container_name: auth-go-servies
    ports:
      - "4000:4000"
    env_file:
      - .env
    depends_on:
      - postgres

  postgres:
    image: postgres:16
    container_name: postgres
    restart: always
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: appdb
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
