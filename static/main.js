var defaultLocation = [59.32, 18.07];
var map = L.map('map', {
    center: defaultLocation,
    zoom: 13
});

navigator.geolocation.getCurrentPosition(function(location) {
  var latlng = new L.LatLng(location.coords.latitude, location.coords.longitude);
  map.flyTo(latlng);
});

var urlParams = new URLSearchParams(window.location.search);
var mapProvider = urlParams.get("provider") || "mapbox";
var mapStyle = urlParams.get("style") || "mapbox/light-v10";

if (mapProvider == "stamen") {
    L.tileLayer("https://stamen-tiles-{s}.a.ssl.fastly.net/{id}/{z}/{x}/{y}.{type}", {
        type: "png", // Watercolor use jpg but png redirects to that.
        id: mapStyle,
        subdomains: ["a", "b", "c", "d"],
        maxZoom: 18,
        attribution: [
            'Map tiles by <a href="https://stamen.com/">Stamen Design</a>, ',
            'under <a href="https://creativecommons.org/licenses/by/3.0">CC BY 3.0</a>. ',
            'Data by <a href="https://openstreetmap.org">OpenStreetMap</a>, ',
            'under <a href="https://www.openstreetmap.org/copyright">ODbL</a>.'
        ].join("")
    }).addTo(map);
} else {
    L.tileLayer('https://api.mapbox.com/styles/v1/{id}/tiles/{z}/{x}/{y}?access_token={accessToken}', {
        attribution: [
            '<a href="https://www.mapbox.com/about/maps/" target="_blank">&copy; Mapbox</a>, ',
            '<a href="http://www.openstreetmap.org/about/" target="_blank">&copy; OpenStreetMap</a> contributors, ',
            '<a href="https://www.mapbox.com/map-feedback/#/-74.5/40/10" target="_blank">Improve this map</a></div>'
        ].join(""),
        maxZoom: 18,
        tileSize: 512,
        zoomOffset: -1,
        id: mapStyle,
        accessToken: 'pk.eyJ1IjoibWF0dGk0cyIsImEiOiJja2E0Nmc3ZXgwYjE3M2ZtdmtpemR5ZHNvIn0.Zd4e3EFiWxz8tFV9MYREbg'
    }).addTo(map);
}

import Athlete from "./Athlete.js";
import QueryForm from "./QueryForm.js";
import ActivityList from "./ActivityList.js";

var app = new Vue({
    el: '#menu',
    data: {
        authed: false,
        activityTypes: {},
        athlete: {},
        activities: [],
    },
    components: {
        Athlete,
        QueryForm,
        ActivityList
    },
    methods: {
        getActivities: function(data) {
            var text = encodeURIComponent(data.queryString);
            fetch(`./activities?q=${text}&after=${data.fromDate}&before=${data.toDate}&type=${data.selectedType}`)
                .then(stream => stream.json())
                .then(addActivityData);
        }
    }
});


fetch("./static/activitytypes.json")
    .then(stream => stream.json())
    .then(data => app.activityTypes = data);

fetch(`./athlete`)
    .then(stream => stream.json())
    .then(function(data) {
        var name = data.firstname + " " + data.lastname;
        app.athlete.name = name;
        app.athlete.url = "https://www.strava.com/athletes/" + data.id;
        app.athlete.image = data.profile_medium;
        app.authed = true;
    });

function addActivityData(data) {
    L.geoJSON(data).bindPopup(function(layer) {
        return "<p>" + layer.feature.properties.name + "</p>";
    }).eachLayer(addToActivityList).addTo(map);
    centerMap();
}

function centerMap() {
    var bounds = L.latLngBounds([]);
    map.eachLayer(function(layer) {
        if (layer.hasOwnProperty("feature")) {
            var layerBounds = layer.getBounds();
            bounds.extend(layerBounds);
        }
    });
    map.flyToBounds(bounds);
}

function addToActivityList(layer) {
    if (layer.hasOwnProperty("feature")) {
        var a = layer.feature.properties.activity;
        a.url = "https://strava.com/activities/" + a.id;
        app.activities.push(a);
    }
}
