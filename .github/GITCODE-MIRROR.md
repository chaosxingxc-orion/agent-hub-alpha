# Agent Hub Alpha cross-platform collaboration

GitHub `chaosxingxc-orion/agent-hub-alpha` is the final integration repository.
GitCode `xingchao_ml/agent-hub-alpha` is the corresponding team workspace.
NGA retains model experiments; this bridge never copies code between the two projects.

GitHub main, github/** branches and v* tags propagate by ordinary fast-forward
pushes. GitCode requests import into gitcode/mr-N; GitHub-native requests project
to github/pr-N. These namespaces are bridge-owned. Developers use other branches.
No automatic approval, merge, force push, branch deletion or tag replacement.
Divergence fails visibly and must be resolved with history preserved.

Comments retain author and source links. Reserved nga-bridge markers prevent
loops and are scoped to this repository pair. Edits update copies; deletions,
labels, assignees, releases, attachments and LFS are outside this bridge.

GitHub events trigger reconciliation. GitCode changes are polled every five
minutes; provider delays mean eventual consistency, not a real-time SLA.
Use workflow_dispatch for immediate catch-up. Runs are serialized, use trusted
main only and never execute imported code. GITCODE_TOKEN is a repository secret;
GitHub uses its per-run token. Never put credentials in comments.

Run both test_sync_gitcode.py and test_collaboration_bridge.py before changing
the bridge. Inspect authentication, divergence or API failures before rerunning.
Do not start a competing reverse full mirror. GitCode-origin requests start as
GitHub drafts; only maintainers change their review readiness. GitHub-origin
requests project their Draft state to GitCode, including on description updates.
Copied reviews and Ready state are not approval or permission to merge.

Main-targeted requests keep main as their target. Other targets use bridge-owned
immutable `github/base-<40-character SHA>` or `gitcode/base-<SHA>` snapshots of the
source base. The description retains the original base name and SHA. When the
source base advances, the open mirror targets its new snapshot; missing or moving
bases fail visibly rather than falling back to main. These snapshots are review
bases, not release branches. Closed requests retain their historical targets.
No developer-owned branch is overwritten. Final integration still occurs only
on GitHub into the source request's declared target.

[Migration and historical discussion](../docs/migration/README.md).
