import Dexie, { type Table } from "dexie";
import type { Activity } from "./types.ts";

export class MamDB extends Dexie {
  activities!: Table<Activity, string>;

  constructor() {
    super("mamDB");
    this.version(1).stores({
      activities: "&id, *text, type, start_date",
    });
  }
}

export const db = new MamDB();
