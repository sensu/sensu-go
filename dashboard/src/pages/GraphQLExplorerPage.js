import React from "react";
import PropTypes from "prop-types";
// import GraphiQL from "graphiql";
import Paper from "material-ui/Paper";
import Typography from "material-ui/Typography";
import { withStyles } from "material-ui/styles";

import RestrictUnauthenticated from "../components/RestrictUnauthenticated";
// import { getAccessToken } from "../utils/authentication";

const styles = theme => ({
  content: {
    margin: theme.spacing.unit * 2,
  },
  explorer: {
    height: 600,
  },
  title: {
    padding: theme.spacing.unit,
    paddingLeft: 14,
  },
});

// function fetchQuery(params) {
//   const parseJson = response => response.json();
//   const makeRequest = accessToken =>
//     fetch("/graphql", {
//       method: "POST",
//       headers: {
//         Accept: "application/json",
//         Authorization: `Bearer ${accessToken}`,
//         "content-type": "application/json",
//       },
//       body: JSON.stringify(params),
//     });
//
//   return getAccessToken()
//     .then(makeRequest)
//     .then(parseJson);
// }

class QueryPage extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
  };

  render() {
    const { classes } = this.props;
    return (
      <RestrictUnauthenticated>
        <Paper className={classes.content}>
          <Typography type="headline" className={classes.title}>
            Sensu GraphQL Explorer
          </Typography>
          <div className={classes.explorer}>
            <p>
              Due to an incompatible GraphQL version, explorer has been removed
              for the time being.
            </p>
            {/* <GraphiQL fetcher={fetchQuery} /> */}
          </div>
        </Paper>
      </RestrictUnauthenticated>
    );
  }
}

export default withStyles(styles)(QueryPage);
