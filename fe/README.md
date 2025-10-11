<a href="https://klyniva.com/">
  <p align="center">
    <picture>
      <img alt="Klyniva Logo" src="/images/klyniva-text.png" width="400">
    </picture>
  </p>
</a>
<p align="center"><b>Our goal is to continuously improve the quality and accessibility of healthcare services using advanced digital tools.</b></p>
<h2></h2>
<h3 align="center"><a href="https://klyniva.com" target="_blank">🚀 Live Deployment</a></h3>
<center>


</center>
<p align="center">Klyniva Healthcare Management System - Advanced tools for modern healthcare delivery.</p>

## Features

- **Comprehensive Patient Management** - Complete patient records, medical history, and care tracking
- **Facility Administration** - Multi-facility support with role-based access control
- **Therapy & Treatment Planning** - Advanced scheduling and treatment management
- **Multi-language Support** - Available in Portuguese, English, and other languages
- **Real-time Monitoring** - Live patient monitoring and vital signs tracking
- **Integrated Timetabling** - Complete scheduling system for healthcare providers

## Sample Data

The system includes sample data for demonstration:
- **Climutua Medical Center** - Featured healthcare facility
- **Sample Patients** - Pre-configured patient records with various conditions
- **Therapy Sessions** - Example treatment schedules and appointments

![Code scanning - action](https://github.com/ohcnetwork/care_fe/workflows/Code%20scanning%20-%20action/badge.svg)
![OSSAR](https://github.com/ohcnetwork/care_fe/workflows/OSSAR/badge.svg)
[![Cypress Tests](https://img.shields.io/endpoint?url=https://cloud.cypress.io/badge/simple/wf7d2m/develop&style=flat&logo=cypress)](https://cloud.cypress.io/projects/wf7d2m/runs)
![Staging Release](https://github.com/ohcnetwork/care_fe/workflows/CARE%20Develop%20Registry/badge.svg)
![Production Release](https://github.com/ohcnetwork/care_fe/workflows/Production%20Release/badge.svg)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/200482ab117e4b5397ff3f5ae5719aa2)](https://www.codacy.com/gh/ohcnetwork/care_fe?utm_source=github.com&utm_medium=referral&utm_content=ohcnetwork/care_fe&utm_campaign=Badge_Grade)
[![Maintainability](https://api.codeclimate.com/v1/badges/f1438f693aa459805301/maintainability)](https://codeclimate.com/github/ohcnetwork/care_fe/maintainability)

## Getting started

- 💬 Comment on the issue if you are willing to take it up, and link the pull request with the issue.
- 🏷️ Tag `@ohcnetwork/care-fe-code-reviewers` for faster resolution.
- 📸 Attach screenshots in the pull requests showing the changes made in the UI.

#### Install the required dependencies

```sh
npm run install-all
```
#### First-time setup

For first-time setup, run the following command to generate the pluginMap and install plugin configurations:

```sh
npm run setup
```

#### 🏃 Run the app in development mode

```sh
npm run dev
```

Once the development server has started, open [localhost:4000](http://localhost:4000) in your browser. The page will be automatically reloaded when you make edits and save. You will also see any lint errors in the console.

#### 🔑 Staging API Credentials

Authenticate to staging API with any of the following credentials

```yaml
- username: dev-districtadmin
  password: Coronasafe@123
  role: District Admin

- username: staffdev
  password: Coronasafe@123
  role: Nurse

- username: doctordev
  password: Coronasafe@123
  role: Doctor
```

#### Contributing to CARE

- Create a branch with branch name of the format `issues/{issue#}/{short-name}` (example `issues/7001/edit-prescriptions`) from the latest [`develop`](https://github.com/ohcnetwork/care_fe/tree/develop) branch when starting to work on an issue.
- Once the changes are pushed to the branch, make a pull request with a meaningful title (example: "💊 Adds support for editing prescriptions" #6369)
- Ensure the issue number is mentioned in the PR with a closing tag by following the PR body template. (Refer: [Linking a pull request to an issue](https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue#linking-a-pull-request-to-an-issue-using-a-keyword))
- Once the code review is done, the PR will be marked with a "Needs Testing" label where it'll be queued for QA testing.
- Once tested, the PR would be marked with a "Tested" label and would be queued for merge.

### Translations

All strings must be encased in i18n translations. New translation strings must be specified in `src`->`Locale`->`en`. Do not add translations for languages other than english through pull requests. Other language translations can be contributed through [Crowdin](https://crowdin.com/project/ohccarefe)

### Testing

To ensure the quality of our pull requests, we use a variety of tools:

- **Automated E2E Testing:** We use Cypress for end-to-end testing to automatically verify the functionality and performance of our code.
- **Manual Real Device Testing:** We use BrowserStack to manually test our code on real devices, ensuring compatibility and functionality across different platforms and browsers.

#### 🧪 Run cypress tests

To run cypress tests locally, you'll need to setup the backend to run locally and load dummy data required for cypress to the database. See [docs](https://github.com/ohcnetwork/care#self-hosting).

Once backend is running locally, you'll have to ensure your local front-end is connected to local backend, by setting the `REACT_CARE_API_URL` env.

```env
#.env
REACT_CARE_API_URL=http://127.0.0.1:9000
```

Once done, start the development server by running

```sh
npm run dev
```

Once development server is running, then run the cypress tests in either of the ways described below.

```sh
npm run cypress:run        # To run all tests in headless mode.
```

```sh
npm run cypress:run:gui    # To run all tests in headed mode.
```

```sh
npm run cypress:open       # To debug and run tests individually.
```

- Failed test screenshots are saved in `cypress/screenshots`
- All test videos are saved in `cypress/videos`

## 📖 Documentations

- [CARE Documentation](https://docs.ohc.network/docs/care)
- [Swagger API Documentation](https://careapi.ohc.network/swagger/)
- [Testing Documentation](https://docs.coronasafe.network/care-testing-documentation/)

## 🚀 Production

#### Build the app for production

```sh
npm run build
```

Builds the app for production to the `build` folder. It correctly bundles React in production mode and optimizes the build for the best performance.

#### Start a production `http-server`

```sh
npm run preview
```

Starts a production http-server in local to run the project with Service worker.
The build is minified and the filenames include the hashes.

**🚀 Your app is ready to be deployed!**
