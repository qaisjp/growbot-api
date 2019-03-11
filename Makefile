PSQL_USER=growbot

.PHONY: schema

reset_schema::
	# kick clients off the database
	psql postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'growbot_dev';"

	# reset schema
	dropdb growbot_dev --if-exists
	createdb growbot_dev
	psql growbot_dev -c 'grant all privileges on database growbot_dev to growbot;'
	cat schema.sql | psql -U ${PSQL_USER} growbot_dev > /dev/null
	@echo "Schema has been reset!"

schema.sql::
	pg_dump -s -U ${PSQL_USER} growbot_dev > schema.sql
	@echo "Schema has been written to file"

# save a copy of dev database into dev_backup
checkpoint::
	mkdir -p dev_backup
	pg_dump -F c -U ${PSQL_USER} growbot_dev > dev_backup/$$(date +%F_%H-%M-%S).dump

# restore latest dev backup
restore_checkpoint::
	dropdb growbot_dev
	createdb growbot_dev
	psql growbot_dev -c 'grant all privileges on database growbot_dev to growbot;'
	pg_restore -U ${PSQL_USER} -d growbot_dev $$(find dev_backup | grep \.dump | sort | tail -n 1)

# save the database entries
save::
	mkdir -p dev_saves
	pg_dump --data-only -F p -v -U ${PSQL_USER} -d growbot_dev -f dev_saves/$$(date +%F_%H-%M-%S).sql
