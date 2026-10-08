WITH arrayMap(value -> toFloat64(value), ['12.5', '13', '0', '14.25']) AS values,
      range(length(values)) AS indices, range(length(values), length(values) + 25) AS future
      , (SELECT recent FROM (
    SELECT index, exponentialTimeDecayedAvg(7)(value, index) OVER (ORDER BY index ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS recent
    FROM (SELECT arrayJoin(arrayZip(values, indices)) AS pair, pair.1 AS value, pair.2 AS index)
  ) ORDER BY index DESC LIMIT 1) AS recent
      SELECT arrayWithConstant(25, recent) AS predictions
