type ChangelogEntry = {
  version: string;
  date: string;
  description: string;
  changes: string[];
};

export const changelog: ChangelogEntry[] = [
  {
    version: "1.5.0",
    date: "2025-09-26",
    description: "Handing over a little more control!",
    changes: [
      "Added a new MinimumBid option to the /config command.",
      "Added a new EnableNationalDex option to the /config command.",
      "Fixed an issue where Pokemon with spaces in their names weren't able to be nominated (for example: Iron Hands).",
    ],
  },
  {
    version: "1.4.0",
    date: "2025-07-18",
    description: "Teams are now easier to keep track of!",
    changes: [
      "Added an embed that displays the current teams, updating after every bid phase.",
    ],
  },
  {
    version: "1.3.0",
    date: "2025-07-16",
    description: "A new QoL feature has arrived!",
    changes: [
      "Created a new command: /pick! Check the GitHub page for more info on how to use the command.",
      "Fixed an issue where bidding more money than you currently have caused the bot to break.",
      "Fixed an issue where bidding below zero would break the bot.",
      "Re-enabled the config command.",
    ],
  },
  {
    version: "1.2.0",
    date: "2025-06-28",
    description: "Per-server configuration is here!",
    changes: [
      "Created a new command: /config! Check the GitHub page for more info on how to use the command.",
    ],
  },
  {
    version: "1.1.3",
    date: "2025-06-26",
    description: "It was me, I dealt it!",
    changes: [
      "Fixed an issue where bidding at the same time as someone else would cause the name of the highest bidder to always be the same person no matter what",
    ],
  },
  {
    version: "1.1.2",
    date: "2025-06-22",
    description:
      "Your bank should probably give proper notice when you spend money... (whoops lol).",
    changes: [
      "Updated the balance notification message to be an embed so that it stands out more",
      "Fixed an issue where the message after the first bidding phase notifying the players of their balances would no longer update after subsequent bidding phases",
      "Fixed an issue where using the /stopall command during a nomination phase didn't edit the message notifying players that the auction was closed",
      `Fixed an issue where using /nominate after using /stopall would display an "Application did not respond" error instead of the proper error message`,
    ],
  },
  {
    version: "1.1.1",
    date: "2025-06-21",
    description: "A quick fix to an annoying bug...",
    changes: [
      `Fixed an issue where if you force start an auction with no players, and then you begin a successful auction, you get an "application did not respond" error after attempting to nominate a Pokemon`,
    ],
  },
  {
    version: "1.1.0",
    date: "2025-06-18",
    description: "The great refactor!",
    changes: [
      "Created a new command: /stopall",
      "Created a new changelog page on the bot website",
      "Refactored the codebase; the bot should run more efficiently now",
    ],
  },
  {
    version: "1.0.0",
    date: "2025-05-20",
    description: "Initial release!",
    changes: [
      "Created 3 new commands: /auction, /nominate, /bid",
      "Created status page",
    ],
  },
];
