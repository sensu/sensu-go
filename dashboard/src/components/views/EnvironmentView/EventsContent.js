import React from "react";
import PropTypes from "prop-types";
import { Query } from "react-apollo";
import gql from "graphql-tag";
import { withStyles } from "@material-ui/core/styles";
import Button from "@material-ui/core/Button";
import AppContent from "/components/AppContent";
import EventsContainer from "/components/EventsContainer";
import SearchBox from "/components/SearchBox";
import Content from "/components/Content";
import NotFoundView from "/components/views/NotFoundView";
import RefreshIcon from "@material-ui/icons/Refresh";
import WithQueryParams from "/components/WithQueryParams";

const styles = theme => ({
  headline: {
    display: "flex",
    alignContent: "center",
    paddingLeft: theme.spacing.unit,
    paddingRight: theme.spacing.unit,
    [theme.breakpoints.up("sm")]: {
      paddingLeft: 0,
      paddingRight: 0,
    },
  },
  searchBox: {
    width: "100%",
    [theme.breakpoints.up("sm")]: {
      width: "50%",
    },
  },
  container: {
    marginTop: 10,
  },
  grow: {
    flex: "1 1 auto",
  },
  hiddenSmall: {
    display: "none",
    [theme.breakpoints.up("sm")]: {
      display: "flex",
    },
  },
});

// If none given default expression is used.
const defaultExpression = "HasCheck && IsIncident";

class EventsContent extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    match: PropTypes.object.isRequired,
  };

  static query = gql`
    query EnvironmentViewEventsContentQuery(
      $filter: String!
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

  render() {
    const { classes, match } = this.props;

    return (
      <WithQueryParams>
        {(query, setQuery) => (
          <Query
            query={EventsContent.query}
            fetchPolicy="cache-and-network"
            variables={{
              ...match.params,
              filter: query.get("filter") || defaultExpression,
              order: query.get("order"),
            }}
          >
            {({ data: { environment } = {}, loading, error, refetch }) => {
              if (error) throw error;
              if (!environment && !loading) return <NotFoundView />;

              return (
                <AppContent>
                  <Content className={classes.headline}>
                    <SearchBox
                      className={classes.searchBox}
                      onSearch={val => setQuery("filter", val)}
                      initialValue={query.get("filter")}
                      placeholder="Filter events…"
                    />
                    <div className={classes.grow} />
                    <Button
                      className={classes.hiddenSmall}
                      onClick={() => refetch()}
                    >
                      <RefreshIcon />
                      reload
                    </Button>
                  </Content>
                  <EventsContainer
                    className={classes.container}
                    onQueryChange={setQuery}
                    environment={environment}
                    loading={loading}
                  />
                </AppContent>
              );
            }}
          </Query>
        )}
      </WithQueryParams>
    );
  }
}

export default withStyles(styles)(EventsContent);
