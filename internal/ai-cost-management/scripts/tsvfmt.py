#!/usr/bin/env python3
import sys
rows=[l.rstrip('\n').split('\t') for l in open(sys.argv[1])]
w=[max(len(r[i]) if i<len(r) else 0 for r in rows) for i in range(max(map(len,rows)))]
for r in rows: print('  '.join(c.ljust(w[i]) for i,c in enumerate(r)))
