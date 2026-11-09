# Notes

A scaffold **copied into the site and edited**. This is `Guestbook` with the command renamed to `note` and each entry stored as `noted: <who> — <text>`. It does not import `scaffold`.

When a merged scaffold is close but not quite right, copy it into your module and change it here. Leave `scaffold/` in the Portico module alone.

```bash
go run .
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
# type: note bring cake
```

See [hello](../hello) for using a scaffold unchanged, and [from-scratch](../from-scratch) for a scene written without one.
