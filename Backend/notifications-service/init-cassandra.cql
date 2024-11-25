CREATE KEYSPACE IF NOT EXISTS trello
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE trello;

CREATE TABLE IF NOT EXISTS notifications_by_month (
    user_id UUID,
    notification_id UUID,
    year_month TEXT,
    created_at TIMESTAMP,
    message TEXT,
    is_read BOOLEAN,
    PRIMARY KEY (user_id, year_month, notification_id)
) WITH CLUSTERING ORDER BY (year_month ASC, notification_id ASC);
