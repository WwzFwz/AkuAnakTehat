package hazard

import "errors"

// MaxEventBytes bounds the compact UTF-8 envelope, including metadata.
// Kafka batches reserve a further MiB for framing and DLQ headers.
const MaxEventBytes = 4 << 20
const MaxPageBytes = 8 << 20

var ErrEventTooLarge = errors.New("event_envelope_too_large")
