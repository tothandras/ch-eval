# Async insert + insert_quorum (insert_quorum_parallel=0): LOGICAL_ERROR on the data-block path

Found by the D7 replica-consistency experiment. It was reproduced with
`experiments/repro/quorum-async-logical-error.sh` on compose 1×2 clusters on
2026-10-08:

```
server: 26.3.33.24
VALUES:       Code: 2
FORMAT TSV:   Code: 49 LOGICAL_ERROR
server: 26.3.33.10001.altinitystable
VALUES:       Code: 2
FORMAT TSV:   Code: 49 LOGICAL_ERROR
```

The combination is invalid, and with `VALUES` the server rejects it with a
proper error: "Async inserts with quorum only make sense with enabled
insert_quorum_parallel setting". When the data is sent as a block, the request
reaches `LOGICAL_ERROR` ("Should be checked earlier") instead. A block is sent
for any `FORMAT` clause, and clickhouse-go's batch inserts always send blocks.

Both builds behave identically, so this is upstream behaviour at 26.3.33, not
an Altinity difference.

Practical consequence for OpenMeter: the strict-consistency policy
(`insert_quorum` + `insert_quorum_parallel=0` + `select_sequential_consistency`)
requires synchronous inserts. Async inserts are the 26.2+ server default.

Not yet checked on newer upstream versions or Cloud. Not reported upstream;
that is left to the owner.
