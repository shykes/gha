#!/bin/bash

# Run a simple query to "warm up" the engine
echo '{directory{id}}' | dagger query
