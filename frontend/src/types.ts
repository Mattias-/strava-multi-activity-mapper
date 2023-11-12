import { Signal } from "@preact/signals";

export interface Athlete {
  id: string;
  firstname: string;
  lastname: string;
  profile_medium: string;
}

export type ActivityRes = {
  description?: string;
  name: string;
  type: string;
  start_date: string;
  id: string;
};

export type Feature = {
  type: string;
  properties: {
    activity: ActivityRes;
  };
};

export interface ApiActivities {
  features: Feature[];
}

export interface Activity {
  feature?: Feature;
  id: string;
  start_date: Date;
  text: string[];
  type: string;
}

export interface AppState {
  activities: Signal<Activity[]>;
}
