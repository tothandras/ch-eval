WITH arrayMap(value -> toFloat64(value), ['12.5', '13', '0', '14.25']) AS values,
      range(length(values)) AS indices, range(length(values), length(values) + 25) AS future
      SELECT arrayMap(day -> arrayAvg(arrayFilter((value, index) -> index % 7 = day % 7, values, indices)), future) AS predictions
