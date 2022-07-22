<p align=right>
<a href=https://autorelease.dmz.general.palantir.tech/palantir/tenablesc-metrics><img src=https://img.shields.io/badge/Perform%20an-Autorelease-success.svg alt=Autorelease></a>
</p>

# Security Center Metrics

Uses the security center API to retrieve data.  Then metrics are calculated and pushed to DataDog.

Can be run either one time (if used in a scheduled context) or continuously.  The provided dockerfile defaults to running once and exiting.

## Configuration

The go struct for the config can be found [here](cmd/config.go#L26)
