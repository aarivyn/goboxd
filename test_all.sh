#!/bin/bash
BASE="http://localhost:8080"

echo "=== healthz ===" 
curl -s $BASE/healthz

echo -e "\n=== readyz ==="
curl -s $BASE/readyz | python3 -m json.tool

echo -e "\n=== Python ==="
curl -s -X POST $BASE/run -H "Content-Type: application/json" \
  -d '{"language":"py3","source":"print(\"hello\")","tests":[{"stdin":"","expected_stdout":"hello"}]}' \
  | python3 -c "import sys,json; r=json.load(sys.stdin); print('✅ PASS' if r['status']=='accepted' else '❌ FAIL: '+r['status'])"

echo "=== C++ ==="
curl -s -X POST $BASE/run -H "Content-Type: application/json" \
  -d '{"language":"cpp","source":"#include<iostream>\nint main(){std::cout<<\"hello\";return 0;}","tests":[{"stdin":"","expected_stdout":"hello"}]}' \
  | python3 -c "import sys,json; r=json.load(sys.stdin); print('✅ PASS' if r['status']=='accepted' else '❌ FAIL: '+r['status'])"

echo "=== C ==="
curl -s -X POST $BASE/run -H "Content-Type: application/json" \
  -d '{"language":"c","source":"#include<stdio.h>\nint main(){printf(\"hello\");return 0;}","tests":[{"stdin":"","expected_stdout":"hello"}]}' \
  | python3 -c "import sys,json; r=json.load(sys.stdin); print('✅ PASS' if r['status']=='accepted' else '❌ FAIL: '+r['status'])"

echo "=== Java ==="
curl -s -X POST $BASE/run -H "Content-Type: application/json" \
  -d '{"language":"java","source":"public class Main{public static void main(String[] args){System.out.print(\"hello\");}}","tests":[{"stdin":"","expected_stdout":"hello"}]}' \
  | python3 -c "import sys,json; r=json.load(sys.stdin); print('✅ PASS' if r['status']=='accepted' else '❌ FAIL: '+r['status'])"

echo "=== Bash ==="
curl -s -X POST $BASE/run -H "Content-Type: application/json" \
  -d '{"language":"bash","source":"echo hello","tests":[{"stdin":"","expected_stdout":"hello"}]}' \
  | python3 -c "import sys,json; r=json.load(sys.stdin); print('✅ PASS' if r['status']=='accepted' else '❌ FAIL: '+r['status'])"

echo "=== JavaScript ==="
curl -s -X POST $BASE/run -H "Content-Type: application/json" \
  -d '{"language":"js","source":"console.log(\"hello\")","tests":[{"stdin":"","expected_stdout":"hello"}]}' \
  | python3 -c "import sys,json; r=json.load(sys.stdin); print('✅ PASS' if r['status']=='accepted' else '❌ FAIL: '+r['status'])"

echo "=== Verilog ==="
curl -s -X POST $BASE/run -H "Content-Type: application/json" \
  -d '{"language":"verilog","source":"module main; initial begin \$display(\"hello\"); \$finish; end endmodule","tests":[{"stdin":"","expected_stdout":"hello"}]}' \
  | python3 -c "import sys,json; r=json.load(sys.stdin); print('✅ PASS' if r['status']=='accepted' else '❌ FAIL: '+r['status'])"

