--!strict

-- The Studio plugin deliberately knows nothing about Roblox OAuth. It only talks
-- to the Lettuce desktop bridge on the loopback interface. The desktop runtime
-- owns token storage, refresh, Roblox requests, uploads, retries, and logging.
return table.freeze({
	VERSION = "0.2.0",
	HOST = "127.0.0.1",
	PORT = 38073,
	POLL_INTERVAL_SECONDS = 0.65,
	MAX_CONSECUTIVE_POLL_FAILURES = 8,
})
