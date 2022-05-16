# Logging Style Guide

## Log messages
* messages should be fixed strings, never interpolate values into them, use fields for values
* message strings should be consistent identification of what action/event was/is happening, consistency makes searching the logs and finding correlated events easier
* error messages should look like any other log messages, no need to say "x failed", the log level and error field are sufficient indication of failure

## Log entry fields
* Create/Use field helpers for commonly used field value types (see logging.go), it promotes consistent format and allows changing it easily in one place.
* Use logging.IfDebug() to include fields in DEBUG mode only, this is useful for values that are usually not interesting or possibly expensive to emit

# Log value helpers
* Make the field creation do as little as possible, i.e. just capture the existing value/object. Postpone any transformation to log emission time by employing generic zap.Stringer, zap.Array, zap.Object fields (see logging.go).

## Logger management
* adorn the logger with fields and reuse the adorned logger rather than repeatedly create fields with each log entry
* prefer passing the adorned logger down the call chain using Context, it promotes consistent log entry structure (fields will exist consistently in related entries)

