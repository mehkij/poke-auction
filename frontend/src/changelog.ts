type ChangelogEntry = {
  version: string;
  date: string;
  description: string;
  changes: string[];
};

export const changelog: ChangelogEntry[] = [
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
