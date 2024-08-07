## README.md

### CalendarFS: A FUSE-based Calendar Filesystem

**CalendarFS** is a Go-based filesystem implementation using the FUSE kernel interface. It presents a calendar as a hierarchical filesystem structure, allowing users to interact with calendar events and appointments through file and directory operations.

#### Features

* **Calendar as Filesystem:** Transforms a calendar into a navigable filesystem.
* **Event-Based Files:** Calendar events are represented as files with metadata.
* **Hierarchical Structure:** Creates a filesystem hierarchy based on calendar categories or dates.
* **FUSE Integration:** Leverages the FUSE kernel interface for filesystem operations.
* **Customizable:** Allows configuration of calendar provider and filesystem structure.

#### Prerequisites

* Go programming language (version 1.18 or later)
* FUSE kernel module
* A calendar provider (e.g., Google Calendar, Outlook, iCalendar)

#### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/AndreRenaud/calfs.git
   cd calfs
   ```
2. **Install dependencies:**
   ```bash
   go mod tidy
   ```
3. **Build the project:**
   ```bash
   go build -o calfs .
   ```

#### Usage

```bash
./calfs --mountpoint <mountpoint> --ics <calender .ics file or URL>
```

* **mountpoint:** The directory where the filesystem will be mounted.
* **ics:** ICS calendar file (either a local file, or a http/https URL)

#### Filesystem Structure

The filesystem structure will look like this:

```
/
├── 2023 
│   ├── 05
│   │   ├── 22
|   │   └── 24
│   └── 07
│       ...
└── 2024
```

#### File Operations

* **Read/Write:** Read or modify event details.

#### Contributing

Contributions are welcome! Please open an issue or submit a pull request.

