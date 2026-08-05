# Valve FM

Una aplicación de radio FM vintage en modo consola para transmitir emisoras desde radio-browser.info.
<img width="3024" height="1898" alt="imagen" src="https://github.com/user-attachments/assets/c9c6ab00-7525-4402-afcd-32ba0e6cd83c" />

## Instalación

### Homebrew (macOS/Linux)

```bash
brew tap zorig/tap
brew install valvefm
```

### Chocolatey (Windows)
> Debido a la revisión de Chocolatey, puede tardar un tiempo en que la nueva versión sea aprobada.

```powershell
choco install valvefm
```

### Desde el código fuente

## Requisitos

- **Go 1.24 o superior**
- **Reproductor de audio:** Reproductor de MP3 puro en Go integrado (sin dependencias externas).
- **Opcional:** `mpv` o `ffplay` para soporte de AAC/OGG y mayor estabilidad en la transmisión.
  - Windows: se descarga automáticamente `ffplay.exe` si es necesario.

## Ejecución (TUI + Bandeja)

```bash
go run ./cmd/radio-tray
```

Notas:

- La aplicación siempre se ejecuta con la bandeja habilitada.
- Ruta del socket en macOS/Linux: `~/.config/valvefm/ctl.sock`
- Archivo de dirección en Windows: `~/.config/valvefm/ctl.addr`
- Windows descarga automáticamente `ffplay.exe` en la primera ejecución si no se encuentra ningún reproductor.

### ⚠️ Advertencia de SmartScreen de Windows
Al ejecutar `valvefm-windows-amd64.exe` por primera vez, Windows podría mostrar una advertencia de "Windows protegió su PC" porque la aplicación no está firmada digitalmente.
1. Haz clic en **Más información**.
2. Haz clic en **Ejecutar de todos modos**.
(Esto es normal para software de código abierto sin un certificado de firma de código costoso.)

### Nota sobre la compilación en Windows

Si deseas que la interfaz de usuario sea visible, compila sin el subsistema gráfico:

```bash
GOOS=windows GOARCH=amd64 go build -o valvefm.exe ./cmd/radio-tray
```

## Atajos de teclado

- Izquierda/Derecha: ajustar sintonizador
- Arriba/Abajo: navegar por emisoras
- [ / ]: cambiar a página anterior/siguiente de emisoras
- Enter: reproducir emisora
- Espacio: pausar/reanudar reproducción
- L: seleccionar país (lista con búsqueda)
- V: mostrar favoritos
- /: buscar emisoras (búsqueda en servidor en modo país, búsqueda local en modo favoritos)
- F: marcar/desmarcar como favorito
- T: cambiar tema
- ?: ayuda
- Q / Ctrl+C: salir

## Notas

- Las emisoras se obtienen de la API de Radio Browser y se ordenan por popularidad.
- La lista de emisoras y los resultados de búsqueda están paginados (200 emisoras por página).
- Si hay favoritos guardados, la aplicación se abre con la lista de favoritos por defecto.
- La selección de país utiliza una lista con búsqueda obtenida de la API.
- Los favoritos se guardan en `~/.config/valvefm/favorites.json`.
- La preferencia de tema se guarda en `~/.config/valvefm/config.json`.
- 12 temas integrados: Vintage, Tokyo Night, Nord, Catppuccin Mocha/Latte, Gruvbox Dark, Dracula, Solarized Dark, One Dark, Rose Pine, Kanagawa, Everforest.

## Lista de verificación de prueba básica

- Inicio: `go run ./cmd/radio-tray` inicia la interfaz de usuario y la bandeja.
- Selector de país: `L` abre la lista, el filtro funciona y Enter carga las emisoras.
- Vista de favoritos: `V` abre los favoritos guardados y también cambia al modo país.
- Reproducción: Enter inicia la reproducción de audio y Espacio pausa/reanuda.
- Siguiente/Anterior: los controles de la bandeja cambian de emisora y reproducen automáticamente.
- Búsqueda: `/` realiza una búsqueda en servidor en modo país y una búsqueda local en modo favoritos.
- Paginación: `[` y `]` permiten moverse entre las páginas de emisoras.
- Salir: la opción Salir de la bandeja y `Q` detienen la reproducción de manera limpia.

## Licencias

Valve FM puede descargar `ffplay.exe` en Windows. Incluya `THIRD_PARTY_NOTICES.md` en su distribución y respete los términos de licencia de FFmpeg.
