#!/usr/bin/env python3
"""探针 stub 闸：只输出 OTHER 形 finding（非 count-claim）。

合同：--check <path> 时向 stdout 打恰一行 `{path}:1: ledger audit: orphan claim`。
用途（m-R1-1 修复钉 FAM-OTHERVER）：证明判定真源=parse_claims(claims)——
把 got 换回旧口径 stdout.count("count claim") 的变异体（A8）会让本 stub 的
段全体掉成 FN/TN，探针即红。import 无副作用（__main__ 守卫）。
"""
import sys

if __name__ == "__main__":
    print(f"{sys.argv[-1]}:1: ledger audit: orphan claim")
