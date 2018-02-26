import React from "react";
import PropTypes from "prop-types";
import GraphiQL from "graphiql";
import Paper from "material-ui/Paper";
import { withStyles } from "material-ui/styles";
import { getAccessToken } from "../utils/authentication";
import AppContent from "../components/AppContent";

require("graphiql/graphiql.css");

const styles = () => ({
  root: {
    maxWidth: "calc(100% - 72px*2)",
  },
  content: {
    flex: "1 1 100%",
    width: "100%",
    height: "100%",
    minHeight: "350px",
    maxHeight: "600px",
  },
});

function fetchQuery(params) {
  const parseJson = response => response.json();
  const makeRequest = accessToken =>
    fetch("/graphql", {
      method: "POST",
      headers: {
        Accept: "application/json",
        Authorization: `Bearer ${accessToken}`,
        "content-type": "application/json",
      },
      body: JSON.stringify(params),
    });

  return getAccessToken()
    .then(makeRequest)
    .then(parseJson);
}

class QueryPage extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
  };

  render() {
    const { classes } = this.props;
    return (
      <AppContent className={classes.root}>
        <Paper className={classes.content}>
          <GraphiQL fetcher={fetchQuery} />
        </Paper>
      </AppContent>
    );
  }
}

export default withStyles(styles)(QueryPage);
