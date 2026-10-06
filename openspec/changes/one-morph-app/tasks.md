## 1. Large modals

- [x] 1.1 Share one modal size, about 96vw by 96dvh, for MorphNotes and AI tools, and remove the 1200px cap on AI tools
- [x] 1.2 Open MorphNotes from the Morph AI header in that modal over the chat, and close it back to the same chat
- [x] 1.3 Load the existing `/morphdata` UI in the MorphNotes modal

## 2. Modules inside MorphNotes

- [x] 2.1 Add Event Logs, Content Maker, and Project to MorphNotes navigation, embedding each existing server
- [x] 2.2 Keep Event Logs on its `/events-info` path and pass the Morph bearer on each embed URL
- [x] 2.3 When one of those origins is down, show a hint and do not mount the iframe
- [x] 2.4 Leave Stories, Stick notes, Timelines, Research, and Generic data in place

## 3. Drop Data Access and the MorphUtils shell

- [x] 3.1 Remove Data Access from every product navigation and do not embed it
- [x] 3.2 Remove the MorphUtils header chip and the MorphUtils mention on the login screen
- [x] 3.3 Stop `./start-all.sh` with no arguments, and `start all`, from starting MorphUtils and Data Access
- [x] 3.4 Update header-link checks so a public MorphUtils URL does not render a chip, and MorphNotes opens the modal

## 4. Project navy

- [x] 4.1 Make the Project header an opaque dark blue and paint selected tabs, the selected project, and the HTML/Markdown pill blue
- [x] 4.2 Change the document preview gradient so it does not start from indigo `#1e1b4b`

## 5. Docs and verification

- [x] 5.1 Update the embed section of `deploy/README.md` so the host is the MorphNotes modal, Data Access is not a product embed, and the cookie failure modes stay
- [x] 5.2 Run the header-link checks and confirm the modal width rule is shared and is not capped at 1200px
