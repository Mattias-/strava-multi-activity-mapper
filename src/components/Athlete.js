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
  template: `
        <div id="athlete">
          <a :href="stravaUrl">
          <img :alt="fullName" class="avatar-img" :src="profileImageUrl">
          <h2>{{ fullName }}</h2>
          </a>
        </div>
`,
};
