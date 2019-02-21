.PHONY: schema

reset_schema::
	dropdb -U postgres growbot_dev --if-exists
	createdb -U postgres growbot_dev
	cat schema.sql | psql -U postgres growbot_dev > /dev/null
	@echo "Schema has been reset!"

schema.sql::
	pg_dump -s -U postgres growbot_dev > schema.sql
	@echo "Schema has been written to file"

# save a copy of dev database into dev_backup
checkpoint::
	mkdir -p dev_backup
	pg_dump -F c -U postgres growbot_dev > dev_backup/$$(date +%F_%H-%M-%S).dump

# restore latest dev backup
restore_checkpoint::
	dropdb -U postgres growbot_dev
	createdb -U postgres growbot_dev
	pg_restore -U postgres -d growbot_dev $$(find dev_backup | grep \.dump | sort | tail -n 1)
