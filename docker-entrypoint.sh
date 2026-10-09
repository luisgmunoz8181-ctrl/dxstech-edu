#!/bin/sh
# Arranca el servidor como usuario no-root.
#
# Si el contenedor inicia como root (caso normal), primero corrige la propiedad
# del volumen de datos —los volúmenes creados por versiones anteriores de la
# imagen pertenecen a root— y luego cede los privilegios con su-exec.
set -e

DATA_DIR="${DATA_DIR:-/app/data}"
APP_UID=10001

if [ "$(id -u)" = "0" ]; then
  mkdir -p "$DATA_DIR"
  # Solo se recorre el volumen cuando su raíz aún no es del usuario de la app.
  if [ "$(stat -c %u "$DATA_DIR")" != "$APP_UID" ]; then
    chown -R app:app "$DATA_DIR"
  fi
  exec su-exec app "$@"
fi

exec "$@"
