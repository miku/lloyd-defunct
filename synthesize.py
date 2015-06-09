#!/usr/bin/env python

import json, random

for i in range(100000000):
    print(json.dumps({"id": random.randint(0, 100000), "value": i}))
