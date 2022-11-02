# Demo of package conditions for PackageRevision

This is a controller that demonstrates how package conditions on the
PackageRevision resource can be used. The controller watches for
PackageRevisions that has a resourceGate with the conditionType "foo".
When it finds one, it will set or update the condition of type "foo" with
status "True" and add a new resourceGate with the condtionType "bar".

1. Install Porch on a cluster.
2. Run the demo controller against the cluster with `make run`.
3. Register a repository with Porch.
4. Create a new PackageRevision with the kpt cli: `kpt alpha rpkg init foo --repository=<repo> -n default --workspace=foo`
5. Add the "foo" readinessGate by editing the PackageRevision: `kubectl edit packagerevisions <packageRevision name>`:
```
spec:
  ...
  readinessGates:
  - conditionType: foo
```
6. See that the controller has added the "foo" conditon and the "bar" readinessGate.


