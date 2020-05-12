var map = L.map('map').setView([59.295457899570465, 18.078887555748224], 13);
L.tileLayer('https://api.mapbox.com/styles/v1/{id}/tiles/{z}/{x}/{y}?access_token={accessToken}', {
    attribution: 'Map data &copy; <a href="https://www.openstreetmap.org/">OpenStreetMap</a> contributors, <a href="https://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="https://www.mapbox.com/">Mapbox</a>',
    maxZoom: 18,
    id: 'mapbox/light-v10',
    accessToken: 'pk.eyJ1IjoibWF0dGk0cyIsImEiOiJja2E0Nmc3ZXgwYjE3M2ZtdmtpemR5ZHNvIn0.Zd4e3EFiWxz8tFV9MYREbg'
}).addTo(map);


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
            fetch(`./activity?q=${text}&after=${data.fromDate}&before=${data.toDate}&type=${data.selectedType}`)
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
    map.fitBounds(bounds);
}

function addToActivityList(layer) {
    if (layer.hasOwnProperty("feature")) {
        var a = layer.feature.properties.activity;
        a.url = "https://strava.com/activities/" + a.id;
        app.activities.push(a);
    }
}
