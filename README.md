[![ci](https://github.com/mehkij/poke-auction/actions/workflows/ci.yml/badge.svg)](https://github.com/mehkij/poke-auction/actions/workflows/ci.yml)

# poke-auction

A Discord bot inspired by the YouTube video made by WolfeyVGC. You can find an invite link for the bot [here](https://pokeauction.vercel.app/). You can also check out the inspiration video [here](https://youtu.be/g_ek_JuSMVo?si=3k_ZY_UPV7eKgMIX)!

## Commands

Currently, the bot has these available commands:

```
/auction [generation (1-9)] [time (in seconds) before the auction begins]
/nominate [pokemon]
/bid [amount]
/stopall
/config [field] [value]
```
### Config

`/config` takes two optional parameters: `field` and `value`. Using the command by itself displays which configuration options you have access to. To change the option, use the command with the field name and the value you want to set it to. Inputting a field name _requires_ that you input a value, and vice versa. So you cannot do: `/config BidTimerDuration` or `/config 30`.

### Pick

`/pick` takes in the name of the Pokemon as a required parameter. This command can only be used when you are the last person to not have a full team of Pokemon.

> [!IMPORTANT]
> The /stopall command closes ALL open auctions in a channel. This command should be used in cases where the bot might break unexpectedly, but this is by no means a catch-all solution! Please notify me of any bot-related bugs and issues that may occur.

## How to Play

Each player begins with 10,000 PokéDollars. If a player happens to run out a money, their team will be filled with [baby pokemon](https://m.bulbapedia.bulbagarden.net/wiki/Baby_Pok%C3%A9mon). After every player has a full team, the bot will generate PokePastes of each player's team that can be pasted into Pokemon Showdown's team builder.

> [!NOTE]
> The PokePaste that is used to import teams into Showdown intentionally does not have any information outside of the Pokemon names. The user is expected to fill in the rest of the information needed (abilities, moves, IVs, etc.)

## Known Issues

This is a list of known issues and bugs with the bot currently. If you come across any bugs that have not already been reported or have not been listed below, please create a GitHub issue detailing the bug along with the steps to reproduce.

- Starting multiple auctions at once causes the auction timers to freeze up.

## Planned Features

No currently planned features as of now! Check back later.

## Contributing

As of right now, I am not accepting any contributions to the bot. This is meant to be a personal project used for my own learning, so it would sort of defeat the purpose in a way.
