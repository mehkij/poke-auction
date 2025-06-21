type ChangelogEntry = {
  version: string;
  date: string;
  description: string;
  changes: string[];
};

export const changelog: ChangelogEntry[] = [
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
