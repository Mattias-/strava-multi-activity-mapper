document.getElementById('to-date').valueAsDate = new Date();

var map = L.map('map').setView([59.295457899570465, 18.078887555748224], 13);
L.tileLayer('https://api.tiles.mapbox.com/v4/{id}/{z}/{x}/{y}.png?access_token=pk.eyJ1IjoibWFwYm94IiwiYSI6ImNpejY4NXVycTA2emYycXBndHRqcmZ3N3gifQ.rJcFIG214AriISLbB6B5aw', {
  maxZoom: 18,
  attribution: 'Map data &copy; <a href="https://www.openstreetmap.org/">OpenStreetMap</a> contributors, ' +
    '<a href="https://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, ' +
    'Imagery © <a href="https://www.mapbox.com/">Mapbox</a>',
  id: 'mapbox.light'
}).addTo(map);


function drawActivities(){
  fetch("./activity?q=%23s")
  .then(function(response) {
    return response.json();
  })
  .then(function(data) {
    var fl = L.geoJSON(data).bindPopup(function (layer) {
      return "<p>" + layer.feature.properties.name + "</p>";
    }).eachLayer(addToActivityList).addTo(map);

      fl.once('load', function (evt) {
    // create a new empty Leaflet bounds object
    var bounds = L.latLngBounds([]);
    // loop through the features returned by the server
    fl.eachFeature(function (layer) {
      // get the bounds of an individual feature
      var layerBounds = layer.getBounds();
      // extend the bounds of the collection to fit the bounds of the new feature
      bounds.extend(layerBounds);
    });

    // once we've looped through all the features, zoom the map to the extent of the collection
    map.fitBounds(bounds);
  });
  });
}

function addToActivityList(layer){
  if( layer.hasOwnProperty("feature")) {
    var a = layer.feature.properties.activity;
    var dateString = new Date(a.start_date_local).toISOString().slice(0, 10);
    document.querySelector("#activity-list tbody").insertAdjacentHTML('beforeend', '<tr><td><a href="https://www.strava.com/activities/'+a.id+'">'+a.name+'</a></td><td>'+dateString+'</td><td>'+ a.type +'</td></tr>');
  }
}


fetch("./athlete")
.then(function(response) {
  return response.json();
})
.then(function(data) {
  var name = data.firstname + " " + data.lastname;
  document.querySelector('#athlete h2').textContent = name;
  document.querySelector('#athlete a').href = "https://www.strava.com/athletes/" + data.id;
  document.querySelector('#athlete .avatar-img').src = data.profile_medium;
  document.querySelector('#athlete .avatar-img').alt = name;
  document.querySelector('.connect').style.display = "none";
  document.querySelector('#athlete').style.display = "block";
  document.querySelector('#settings').style.display = "block";
});


