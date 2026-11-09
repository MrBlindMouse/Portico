# From scratch

Reference consumer inside the Portico module. Your site is a separate module (`portico init -style scratch`, or the README “Or write it” steps). Do not add your scenes under `examples/`.

Smallest custom experience: one scene, one action, local flash via `SetLocalEffect`. No scaffolds.

Same program as the README “Or write it” example, and as `portico init -style scratch`.

```bash
go run .
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
# type: greet
```

For an unchanged scaffold see [hello](../hello). For a copied and edited one see [notes](../notes). Guide: [docs/building-a-site.md](../../docs/building-a-site.md).
