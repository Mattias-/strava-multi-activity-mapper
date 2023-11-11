import {
  MapContainer,
  TileLayer,
  GeoJSON,
  Popup,
  ZoomControl,
} from "react-leaflet";
import { LatLngExpression } from "leaflet";
import "leaflet/dist/leaflet.css";
import { AppState, Activity } from "./types.ts";

type MapProps = {
  state: AppState;
};

export default function Map({ state }: MapProps) {
  const center: LatLngExpression = [59.32, 18.07];
  const accessToken =
    "pk.eyJ1IjoibWF0dGk0cyIsImEiOiJja2E0Nmc3ZXgwYjE3M2ZtdmtpemR5ZHNvIn0.Zd4e3EFiWxz8tFV9MYREbg";
  const url = `https://api.mapbox.com/styles/v1/mapbox/light-v10/tiles/{z}/{x}/{y}?access_token=${accessToken}`;
  const attribution = [
    '<a href="https://www.mapbox.com/about/maps/" target="_blank" rel="noreferrer">&copy; Mapbox</a>, ',
    '<a href="http://www.openstreetmap.org/about/" target="_blank" rel="noreferrer">&copy; OpenStreetMap</a> contributors, ',
    '<a href="https://www.mapbox.com/map-feedback/" target="_blank" rel="noreferrer">Improve this map</a></div>',
  ].join("");

  const features = state.activities.value.map((a: Activity) => {
    if (a.feature == undefined) {
      return;
    }
    return (
      <GeoJSON data={a.feature}>
        <Popup>
          <a href={`https://strava.com/activities/${a.id}`}>
            {a.feature.properties.activity.name}
          </a>
          &nbsp;{a.feature.properties.activity.start_date}
        </Popup>
      </GeoJSON>
    );
  });

  return (
    <MapContainer
      className={"w-screen h-screen"}
      center={center}
      zoom={13}
      scrollWheelZoom={true}
      zoomControl={false}
    >
      <ZoomControl position="topright" />

      <TileLayer
        attribution={attribution}
        url={url}
        maxZoom={18}
        tileSize={512}
        zoomOffset={-1}
      />
      {features}
    </MapContainer>
  );
}
