# poke-auction

A Discord bot inspired by the YouTube video made by WolfeyVGC. You can check out the video [here](https://youtu.be/g_ek_JuSMVo?si=3k_ZY_UPV7eKgMIX)!

## Commands

Currently, the bot has these available commands:

```
/auction [generation] [time]
/nominate [pokemon]
/bid [amount]
/stopall
```

> [!IMPORTANT]
> The /stopall command closes ALL open auctions in a channel. This command should be used in cases where the bot might break unexpectedly, but this is by no means a catch-all solution! Please notify me of any bot-related bugs and issues that may occur.

## How to Play

Each player begins with 10,000 PokéDollars. If a player happens to run out a money, their team will be filled with [baby pokemon](https://m.bulbapedia.bulbagarden.net/wiki/Baby_Pok%C3%A9mon). After every player has a full team, the bot will generate PokePastes of each player's team that can be pasted into Pokemon Showdown's team builder.

> [!NOTE]
> The PokePaste that is used to import teams into Showdown intentionally does not have any information outside of the Pokemon names. The user is expected to fill in the rest of the information needed (abilities, moves, IVs, etc.)

## Known Issues

This is a list of known issues and bugs with the bot currently. If you come across any bugs that have not already been reported or have not been listed below, please create a GitHub issue detailing the bug along with the steps to reproduce.

- Starting multiple auctions at once causes the auction timers to freeze up.
- There seems to be a bug where if an auction fails due to no players joining, trying to nominate a Pokemon in a subsequently successful auction causes an index out of range error, resulting in an "application did not respond" message.
  - Something also happens where if you try for the third time to start an auction, it goes through, but the message that gets edited is the first nomination message, not the most recent one.

## Planned Features

- Per-server bot configuration (custom bid timers, custom starting money amounts, etc.)

## Contributing

As of right now, I am not accepting any contributions to the bot. This is meant to be a personal project used for my own learning, so it would sort of defeat the purpose in a way.
