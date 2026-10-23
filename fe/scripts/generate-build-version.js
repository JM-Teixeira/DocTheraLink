#!/usr/bin/env node

const fs = require("fs");
const { v4: uuidv4 } = require("uuid");

const appVersion = uuidv4();

// version: a new id per build (the "check for update" button compares it);
// built_at: when this build was made (shown on the About page)
const jsonData = {
  version: appVersion,
  built_at: new Date().toISOString(),
};

const jsonContent = JSON.stringify(jsonData);

fs.writeFile(
  "./public/build-meta.json",
  jsonContent,
  { flag: "w+", encoding: "utf8" },
  (err) => {
    if (err) {
      return console.log(err);
    }
    return null;
  },
);
