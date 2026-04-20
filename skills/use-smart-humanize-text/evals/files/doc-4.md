# Rate Limits

Rate limits are a pivotal mechanism for ensuring the long-term health of our platform. We've meticulously engineered our rate-limit system to gracefully handle traffic spikes while fostering fair access for every customer — regardless of plan tier.

The default limits are 1000 requests per minute per API key, with comprehensive headers on every response describing your remaining budget. To utilize our system at scale, we recommend implementing exponential backoff with jitter. This approach leverages the well-known properties of randomized retry schedules to streamline recovery and unlock consistent throughput under load, empowering you to ship resilient integrations.
