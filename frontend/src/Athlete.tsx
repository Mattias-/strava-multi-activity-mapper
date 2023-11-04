import { useEffect, useState } from "preact/hooks";
import { Athlete } from "./types.ts";
import { getAthlete } from "./api.ts";

function fullName(athlete: Athlete) {
  return `${athlete?.firstname} ${athlete?.lastname}` || "";
}

function athleteUrl(athlete: Athlete) {
  return `https://www.strava.com/athletes/${athlete?.id}` || "";
}

export default function AthleteComponent() {
  const [isLoading, setIsLoading] = useState(true);
  const [athlete, setAthlete] = useState<Athlete | null>(null);

  useEffect(() => {
    (async () => {
      const a = await getAthlete().catch(console.error);
      if (a !== undefined) {
        setAthlete(a);
      }
      setIsLoading(false);
    })();
  }, []);

  if (isLoading) {
    return <></>;
  }
  if (athlete === null) {
    return (
      <div class="flex shadow-md rounded px-8 pt-4 pb-4 mb-4 max-w-xs bg-gray-100 min-h-32 max-h-32">
        <a href="http://localhost:3000/auth" class="menu-image">
          <img src="/connectwith_strava.svg" alt="Connect with Strava" />
        </a>
      </div>
    );
  }
  return (
    <div class="flex shadow-md rounded px-8 pt-4 pb-4 mb-4 max-w-xs bg-gray-100 items-center min-h-32 max-h-32">
      <a href={athleteUrl(athlete)}>
        <img
          alt={fullName(athlete)}
          class="w-10 h-10 rounded-full mr-4"
          src={athlete?.profile_medium}
        />
      </a>
      <p class="text-gray-900 text-sm leading-none">{fullName(athlete)}</p>
    </div>
  );
}
