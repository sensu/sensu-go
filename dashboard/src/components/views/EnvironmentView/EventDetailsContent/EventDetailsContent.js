import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { Query } from "react-apollo";
import Grid from "material-ui/Grid";

import AppContent from "/components/AppContent";
import Content from "/components/Content";
import NotFoundView from "/components/views/NotFoundView";
import Loader from "/components/Loader";
import EventDetailsCheckResult from "./EventDetailsCheckResult";
import EventDetailsRelatedEntities from "./EventDetailsRelatedEntities";
import EventDetailsConfiguration from "./EventDetailsConfiguration";

const query = gql`
  query EventDetailsContentQuery(
    $ns: NamespaceInput!
    $check: String!
    $entity: String!
  ) {
    event(ns: $ns, entity: $entity, check: $check) {
      id
      timestamp

      check {
        ...EventDetailsCheckResult_check
        ...EventDetailsConfiguration_check
      }
      entity {
        ...EventDetailsCheckResult_entity
        ...EventDetailsRelatedEntities_entity
        ...EventDetailsConfiguration_entity
      }
    }
  }

  ${EventDetailsCheckResult.fragments.check}
  ${EventDetailsCheckResult.fragments.entity}
  ${EventDetailsRelatedEntities.fragments.entity}
  ${EventDetailsConfiguration.fragments.check}
  ${EventDetailsConfiguration.fragments.entity}
`;

class EventDetailsContent extends React.PureComponent {
  static propTypes = {
    match: PropTypes.object.isRequired,
  };

  render() {
    const { match } = this.props;
    const ns = {
      organization: match.params.organization,
      environment: match.params.environment,
    };

    return (
      <Query
        query={query}
        fetchPolicy="cache-and-network"
        variables={{ ...match.params, ns }}
      >
        {({ data: { event } = {}, loading }) => {
          if (!loading && !event) return <NotFoundView />;

          return (
            <AppContent>
              <Content>
                <Grid container>
                  <Loader loading={loading} passthrough>
                    {event && (
                      <React.Fragment>
                        <Grid item xs={12}>
                          <EventDetailsCheckResult
                            check={event.check}
                            entity={event.entity}
                          />
                        </Grid>
                        <Grid item xs={12} md={6}>
                          <EventDetailsRelatedEntities entity={event.entity} />
                        </Grid>
                        <Grid item xs={12} md={6}>
                          <EventDetailsConfiguration
                            check={event.check}
                            entity={event.entity}
                          />
                        </Grid>
                      </React.Fragment>
                    )}
                  </Loader>
                </Grid>
              </Content>
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default EventDetailsContent;
