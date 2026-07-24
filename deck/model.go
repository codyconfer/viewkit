package deck

// Model is the stateful Bubble Tea program root for deck: navigable view
// stack, injectable chrome, and async status strip. It implements tea.Model
// (Init / Update / View).
//
// Contract:
//   - One Model owns one tea.Program for a deck session.
//   - Views receive *Model in Update so they can Push/Pop and read size.
//   - Domain state lives in View implementations (or app kits), not in Model.
//
// Host is a compatibility alias; prefer Model in new code and docs.
type Model = Host
