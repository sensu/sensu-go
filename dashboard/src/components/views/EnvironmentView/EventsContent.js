import React from "react";
import PropTypes from "prop-types";
import { Query } from "react-apollo";
import gql from "graphql-tag";
import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";

import AppContent from "../../AppContent";
import EventsContainer from "../../EventsContainer";
import SearchBox from "../../SearchBox";

import NotFoundView from "../../views/NotFoundView";

// If none given default expression is used.
const defaultExpression = "HasCheck && IsIncident";

class EventsContent extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    history: PropTypes.object.isRequired,
    match: PropTypes.object.isRequired,
    location: PropTypes.object.isRequired,
  };

  static styles = theme => ({
    headline: {
      display: "flex",
      justifyContent: "space-between",
      alignContent: "center",
    },
    searchBox: {
      width: "100%",
      marginLeft: theme.spacing.unit,
      marginRight: theme.spacing.unit,
      [theme.breakpoints.up("sm")]: {
        width: "auto",
        margin: 0,
      },
    },
    title: {
      alignSelf: "flex-end",
      display: "none",
      [theme.breakpoints.up("sm")]: {
        display: "flex",
      },
    },
    container: {
      marginTop: 10,
    },
  });

  static query = gql`
    query EnvironmentViewEventsContentQuery(
      $filter: String = "${defaultExpression}"
      $order: EventsListOrder = SEVERITY
      $environment: String!
      $organization: String!
    ) {
      environment(organization: $organization, environment: $environment) {
        ...EventsContainer_environment
      }
    }

    ${EventsContainer.fragments.environment}
  `;

  constructor(props) {
    super(props);

    const query = new URLSearchParams(props.location.search);

    let filterValue = query.filter;
    if (filterValue === undefined) {
      filterValue = defaultExpression;
    }
    this.state = { filterValue };
  }

  changeQuery = (key, val) => {
    const { location, history } = this.props;
    const query = new URLSearchParams(location.search);

    if (key === "filter") {
      this.setState({ filterValue: val });
    }
    query.set(key, val);

    history.push(`${location.pathname}?${query.toString()}`);
  };

  requerySearchBox = filterValue => {
    if (filterValue.length >= 10) {
      this.changeQuery("filter", filterValue);
    }
    this.setState({ filterValue });
  };

  render() {
    const { classes, match, location, ...props } = this.props;
    const query = new URLSearchParams(location.search);
    return (
      <Query
        query={EventsContent.query}
        variables={{ ...match.params, filter: query.get("filter") }}
        // TODO: Replace polling with query subscription
        pollInterval={5000}
      >
        {({ data: { environment } = {}, loading, error }) => {
          // TODO: Connect this error handler to display a blocking error alert
          if (error) throw error;

          if (!environment && !loading) return <NotFoundView />;

          return (
            <AppContent>
              <div>
                <div className={classes.headline}>
                  <Typography className={classes.title} variant="headline">
                    Events
                  </Typography>
                  <SearchBox
                    className={classes.searchBox}
                    onChange={this.requerySearchBox}
                    value={this.state.filterValue}
                  />
                </div>
                {environment ? (
                  <EventsContainer
                    className={classes.container}
                    onQueryChange={this.changeQuery}
                    environment={environment}
                    {...props}
                  />
                ) : (
                  <div>Loading...</div>
                )}
              </div>
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default withStyles(EventsContent.styles)(EventsContent);
