<script>
export default {
  props: ["athlete"],
  computed: {
    stravaUrl() {
      return "https://www.strava.com/athletes/" + this.athlete.id;
    },
    fullName() {
      return this.athlete.firstname + " " + this.athlete.lastname;
    },
    profileImageUrl() {
      return this.athlete.profile_medium;
    },
  },
};
</script>

<template>
  <div id="athlete">
    <a :href="stravaUrl">
    <img :alt="fullName" class="avatar-img" :src="profileImageUrl">
    <h2>{{ fullName }}</h2>
    </a>
  </div>
</template>

<style>
#athlete {
  text-align: center;
}

#athlete a {
  text-decoration: none;
  color: #000;
}

#athlete h2 {
  font-size: 22px;
}

#athlete img.avatar-img {
  border-radius: 50%;
  vertical-align: middle;
}
</style>
