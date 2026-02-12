# Auto-deploy VPS (Backend)

Este proyecto tiene auto-deploy por branch en [ .github/workflows/deploy-vps.yml ](.github/workflows/deploy-vps.yml):

- Push a `develop` → despliega en producción (`/opt/neighborhood-api/main`) y reinicia `neighborhood-api`
- Push a `main` → despliega en producción (`/opt/neighborhood-api/main`) y reinicia `neighborhood-api`

## 1) Secrets en GitHub

Configura estos secrets en el repositorio `neighborhood-api`:

- `VPS_HOST` (IP o dominio)
- `VPS_USER` (usuario SSH)
- `VPS_PORT` (ej. `22`)

Autenticación (modo híbrido):

- Recomendado: `VPS_SSH_KEY` (llave privada completa)
- Fallback: `VPS_PASSWORD` (si no defines key)

## 2) Preparación inicial del VPS

```bash
sudo mkdir -p /opt/neighborhood-api/main/bin
sudo mkdir -p /opt/neighborhood-api/main/config
```

Crea los archivos de entorno (no se suben desde GitHub Actions):

- `/opt/neighborhood-api/main/config/.env`

## 3) Services systemd

### Producción (`/etc/systemd/system/neighborhood-api.service`)

```ini
[Unit]
Description=Neighborhood API (production)
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/neighborhood-api/main
EnvironmentFile=/opt/neighborhood-api/main/config/.env
ExecStart=/opt/neighborhood-api/main/bin/current
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
```

Recarga y habilita:

```bash
sudo systemctl daemon-reload
sudo systemctl enable neighborhood-api
sudo systemctl start neighborhood-api
```

## 4) Flujo de deploy

- Haz push a `develop` o `main` y ambas desplegarán a producción.

Logs:

```bash
sudo journalctl -u neighborhood-api -f
```
