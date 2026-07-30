# What This Project Actually Is — Explained Simply

No jargon. If you read this top to bottom, you'll understand what this code does
and why it exists.

---

## First: what is CBDC?

**CBDC = Central Bank Digital Currency.** It's digital cash issued by a country's
central bank.

Not Bitcoin. Not the balance in your bank app. It's real government money that
happens to be digital — the same rupee, just in electronic form instead of paper.

India has one (the digital rupee, e₹). The UAE has one (the digital dirham).
Dozens of countries are building them.

---

## The problem this project solves

Imagine you run a bank. You want to let a customer in India send money to a
supplier in Dubai, using digital rupees on one end and digital dirhams on the
other.

Here's the annoying part: **every country built their digital currency
differently.**

- India's system might work one way — say, you send a request in one format, and
  errors come back saying `"ERR_NO_FUNDS"`.
- The UAE's system works completely differently — different format, different
  login method, and its errors say `"insufficient_balance_error"`.
- Gift City's is different again.

They all do roughly the same *things* — create money, move money, destroy money —
but each one speaks its own language.

**So your banking software would need to learn every language separately.** And
every time a new country launches a digital currency, you'd write all that code
again from scratch. Ten countries, ten completely separate chunks of code, ten
things to maintain and ten things that can break.

---

## The solution: a universal adapter

Think of a **universal travel power adapter.**

Wall sockets are different in every country — UK, India, US, Europe, all
different shapes. But your laptop only has one plug. So instead of buying a new
laptop for every country, you carry one adapter. Your laptop plugs into the
adapter, and the adapter handles the country-specific weirdness.

**This project is that adapter, for digital currencies.**

```
                      ┌──────────────────┐
   Your banking       │                  │──→ India's digital rupee
   software    ──────→│   THIS PROJECT   │──→ UAE's digital dirham
   (Fineract)         │  (the adapter)   │──→ Gift City's system
                      │                  │──→ (any future country)
                      └──────────────────┘
```

Your banking software learns **one** way of asking for things. This project
handles translating that into whatever each country actually wants.

Add a new country later? You write one small new piece that plugs into the
adapter. **Nothing in your main banking software has to change.** That's the
whole point.

---

## What is Fineract, and what does it "lack"?

**Fineract** is free, open-source banking software. Real banks and microfinance
institutions use it to run accounts, loans, and payments. Think of it as the
engine room of a bank.

Fineract was built for **regular money** — deposits, withdrawals, loans, ordinary
transfers.

It was **not** built for digital currencies, because those didn't exist when it
was designed.

So this project isn't fixing something broken in Fineract. It's **adding a new
capability** that was never there. Like a car that works perfectly but only takes
petrol — this is the kit that lets it also run on electricity.

---

## What can it actually do?

The project defines **five things you can do with digital money.** These are the
complete lifecycle — money is born, moves around, and dies.

| Action | What it means | Everyday comparison |
|---|---|---|
| **Issue** | Create new digital money | The central bank printing notes |
| **Transfer** | Move money from one person to another | Sending someone cash |
| **Lock** | Freeze money temporarily so it can't be spent | A hotel putting a hold on your card |
| **Burn** | Permanently destroy digital money | Shredding old banknotes |
| **Redeem** | Turn digital money back into regular money | Cashing in a gift card |

Plus a few "just checking" actions that don't move any money:

- **Check balance** — how much is in this wallet?
- **Look up a transaction** — what happened with payment #12345?
- **Check status** — did that payment go through yet?
- **Health check** — is the connection to India's system working right now?

---

## The most important one: "Lock"

This is worth explaining on its own, because it's the reason cross-border
payments are hard.

**The problem:** Sending money from India to Dubai isn't one action — it's two.
Rupees have to leave one account, and dirhams have to arrive in another. Two
separate systems, in two separate countries.

**What could go wrong:** The rupees leave... and then the dirham side fails.

Now the money has vanished. It's gone from India and never arrived in Dubai.
That's a catastrophe.

**The fix — "Lock":** Before anything moves, you freeze the money on both sides.
Neither side can spend it. *Then* you check that both sides are ready. Only when
both are confirmed do you release everything at once.

If anything fails partway through, the locks simply expire and everyone's money
stays exactly where it was. Nothing is lost.

This is the same idea as **escrow** when buying a house. Your money sits in a
holding account, untouchable, until the paperwork clears. If the deal collapses,
you get your money back — it never went anywhere risky.

**Without "Lock", safe cross-border payments are basically impossible.** This is
why it's built into the core of this project rather than added on later.

---

## What's actually in this folder?

About 20 files. Here's what each group does, in plain terms:

### The rulebook (the most important file)

One file lists **every action a connector must be able to perform.** It's a
checklist, not working code — it says *what* must be possible, not *how*.

Like a job description. "The person in this role must be able to do X, Y, and Z."
It doesn't say how they should do it. Anyone who can tick every box qualifies.

So when someone builds the India connector, they must handle all five money
actions plus the status checks. If they miss one, it doesn't qualify — the
computer catches it immediately, before anything ships. **That's the real value:
it makes it impossible to build a half-finished connector by accident.**

### The money counter

A dedicated file just for handling amounts of money. This sounds boring. It isn't.

Computers are famously bad at decimals. Ask most programming languages what
`0.1 + 0.2` equals and you get `0.30000000000000004`. Harmless when you're
measuring a photo's width. **Catastrophic when you're moving someone's salary
across a border.**

So this file never uses decimals. It stores ₹1000.00 as **100000 paise** — a
whole number — and remembers to put the decimal point back when displaying it.

Whole numbers are always exact. Nothing ever drifts. Do a billion transactions
and not one paisa goes missing.

This is one of the genuinely well-made parts of the project.

### The problem list

A catalogue of everything that can go wrong, with a standard name for each:
*not enough money*, *wallet doesn't exist*, *the network is down*, *this failed a
regulatory check*, *the freeze expired*, and about 25 more.

The point is **consistency.** India's system, the UAE's system, and every future
system all report failures using these same names. So the software above only
has to learn one set of problems, not one set per country.

### The settings file

All the adjustable knobs, in one place: which web address to connect to, the
password, how long to wait before giving up, how many times to retry, how fast to
send requests.

Nothing is hardcoded. To point at a different country's test system, you change a
settings file. You don't touch the code.

### The watchers (two files)

- One counts things: how many payments went through, how many failed, how long
  each took. Feeds the dashboards the operations team stares at.
- One tags every payment with a tracking number, so if payment #4471 goes wrong
  you can trace its exact path through every system it touched.

Like parcel tracking, but for money.

### The inspector

Checks that requests make sense **before** sending them anywhere. Is the amount
positive? Is the currency code real? Are the sender and receiver actually
different people?

Catching a mistake here costs nothing. Catching it after money has moved is a
very expensive phone call.

### The pretend network (the biggest single file)

A **fake** digital currency network that lives entirely inside the software.

You obviously can't test against India's real central bank system whenever a
developer changes a line of code. So this fakes it. It behaves like a real
network — balances go up and down, payments succeed — but no real money is
involved.

The clever part: **you can order it to fail on purpose.**

- "Pretend the network is down."
- "Pretend this payment takes 10 seconds."
- "Pretend this wallet has no money."

That way you can test what your software does when things go wrong — *without
waiting for things to actually go wrong in production.* You're rehearsing the
fire drill instead of waiting for a fire.

### The examples and instructions

The rest are worked examples and documentation showing how to use all of the
above.

---

## Is it finished?

**The thinking: yes, and it's good.** The design is solid. The money handling is
correct. The "Lock" mechanism is properly thought through. The list of things
that can go wrong is thorough.

**The code: not yet.** Two honest caveats:

**1. It won't run in its current form.**

The files are named like a numbered report — `1.go.mod.go`, `2.connector.go.go`,
`3.models.go.go`. That's how you organise a *document*, not a working program.
The programming language this is written in has strict rules about folder
structure, and this doesn't follow them. There are also a handful of small
mistakes of the kind you'd catch in five minutes the first time you tried to run
it — the sort that only appear when code is *written out* rather than *built*.

This looks like a **design deliverable** — code written to be reviewed and
approved, not yet code that's been assembled and switched on. The folder is
labelled "Production-Ready," which is optimistic.

**2. Three real things are missing.**

- **There's no way to un-freeze money early.** You can lock it, but the only way
  to release it is to wait for the timer to run out. If a payment gets cancelled,
  the money stays stuck until then. That needs an "unlock" action.
- **It can only ask, never be told.** To find out if a payment succeeded, your
  software has to keep asking "done yet? done yet?" There's no way for the
  network to just call back and say "it's done." The groundwork for this was
  started, but never connected up.
- **Every country's connector will duplicate the same retry logic.** The settings
  file has options for "retry failed requests" and "don't send too fast" — but
  nothing in this project actually *does* the retrying. So whoever builds the
  India connector writes that logic, and the UAE connector writes it again, and
  so on. Since avoiding that exact duplication is the whole point of the project,
  it should be built once, here.

---

## The one-paragraph summary

Different countries built their digital currencies in incompatible ways. Without
this project, banking software would need separate custom code for each one —
expensive to build, worse to maintain, and it gets harder with every new country.
This project is a universal adapter: your banking software learns one way of
asking for things, and the adapter handles each country's quirks. It defines
exactly what "sending digital money" means, how to count money without ever
losing a paisa, what to call each thing that can go wrong, and — most importantly
— how to freeze money on both sides of a border so a cross-border payment can
never half-complete and lose someone's cash. It's the foundation the rest of the
platform gets built on, which is why getting it right matters more than getting
it fast.
