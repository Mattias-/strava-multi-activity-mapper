<style src="leaflet/dist/leaflet.css"></style>

<script>
import * as L from 'leaflet';

export default {
  data() {
    return {
      mapObject: {},
    };
  },
  methods: {
    addActivityData: function (data) {
      L.geoJSON(data)
        .bindPopup(function (layer) {
          var props = layer.feature.properties;
          return `<a href="https://strava.com/activities/${props.activity.id}">${props.name}</a> ${props.activity.start_date_local}`;
        })
        .addTo(this.mapObject);
    },
    centerMap: function () {
      var bounds = L.latLngBounds([]);
      this.mapObject.eachLayer(function (layer) {
        if (Object.prototype.hasOwnProperty.call(layer, "feature")) {
          var layerBounds = layer.getBounds();
          bounds.extend(layerBounds);
        }
      });

      if (bounds.isValid()) {
        this.mapObject.flyToBounds(bounds);
      }
    },
  },
  mounted() {
    var mapObject = (this.mapObject = L.map("map", {
      center: [59.32, 18.07],
      zoom: 13,
    }));
    mapObject.on("locationfound", function () {
      mapObject.setZoom(13);
    });
    mapObject.locate({ setView: true, maxZoom: 10 });

    var urlParams = new URLSearchParams(window.location.search);
    var mapProvider = urlParams.get("provider") || "mapbox";
    var mapStyle = urlParams.get("style") || "mapbox/light-v10";

    if (mapProvider == "stamen") {
      L.tileLayer(
        "https://stamen-tiles-{s}.a.ssl.fastly.net/{id}/{z}/{x}/{y}.{type}",
        {
          type: "png", // Watercolor use jpg but png redirects to that.
          id: mapStyle,
          subdomains: ["a", "b", "c", "d"],
          maxZoom: 18,
          attribution: [
            'Map tiles by <a href="https://stamen.com/">Stamen Design</a>, ',
            'under <a href="https://creativecommons.org/licenses/by/3.0">CC BY 3.0</a>. ',
            'Data by <a href="https://openstreetmap.org">OpenStreetMap</a>, ',
            'under <a href="https://www.openstreetmap.org/copyright">ODbL</a>.',
          ].join(""),
        }
      ).addTo(this.mapObject);
    } else {
      L.tileLayer(
        "https://api.mapbox.com/styles/v1/{id}/tiles/{z}/{x}/{y}?access_token={accessToken}",
        {
          attribution: [
            '<a href="https://www.mapbox.com/about/maps/" target="_blank" rel="noreferrer">&copy; Mapbox</a>, ',
            '<a href="http://www.openstreetmap.org/about/" target="_blank" rel="noreferrer">&copy; OpenStreetMap</a> contributors, ',
            '<a href="https://www.mapbox.com/map-feedback/#/-74.5/40/10" target="_blank" rel="noreferrer">Improve this map</a></div>',
          ].join(""),
          maxZoom: 18,
          tileSize: 512,
          zoomOffset: -1,
          id: mapStyle,
          accessToken:
            "pk.eyJ1IjoibWF0dGk0cyIsImEiOiJja2E0Nmc3ZXgwYjE3M2ZtdmtpemR5ZHNvIn0.Zd4e3EFiWxz8tFV9MYREbg",
        }
      ).addTo(this.mapObject);
    }
  },
};
</script>

<template>
  <div id="map"></div>
</template>

<style>
#map {
  width: 100%;
  height: 100%;
}
</style>
