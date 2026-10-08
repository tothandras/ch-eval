WITH arrayMap(value -> toFloat64(value), ['12.5', '13', '0', '14.25']) AS values,
      range(length(values)) AS indices, range(length(values), length(values) + 25) AS future
      , arrayReduce('simpleLinearRegression', indices, values) AS fit
      SELECT arrayMap(day -> fit.1 * day + fit.2, future) AS predictions
