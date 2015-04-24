#!/usr/bin/env python

import json, random

for i in range(10000000):
    print(json.dumps({"id": random.randint(0, 10000), "value": i}))
