import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import CollapsingMenu from "/components/partials/CollapsingMenu";
import ModalController from "/components/controller/ModalController";
import ListToolbar from "/components/partials/ListToolbar";
import LiveIcon from "/icons/Live";
import NotFoundView from "/components/views/NotFoundView";
import PlusIcon from "@material-ui/icons/Add";
import Query from "/components/util/Query";
import SearchBox from "/components/SearchBox";
import SilencesList from "/components/partials/SilencesList";
import SilenceEntryDialog from "/components/partials/SilenceEntryDialog";
import { withQueryParams } from "/components/QueryParams";
import AppLayout from "/components/AppLayout";

// duration used when polling is enabled; set fairly high until we understand
// the impact.
const pollInterval = 2500; // 2.5s

class SilencesContent extends React.Component {
  static propTypes = {
    match: PropTypes.object.isRequired,
    queryParams: PropTypes.shape({
      offset: PropTypes.string,
      limit: PropTypes.string,
    }).isRequired,
    setQueryParams: PropTypes.func.isRequired,
  };

  static query = gql`
    query EnvironmentViewSilencesContentQuery(
      $environment: String!
      $organization: String!
      $limit: Int
      $offset: Int
      $order: SilencesListOrder
      $filter: String
    ) {
      environment(organization: $organization, environment: $environment) {
        ...SilencesList_environment
      }
    }

    ${SilencesList.fragments.environment}
  `;

  render() {
    const { match, queryParams, setQueryParams } = this.props;
    const { limit = "25", offset = "0", order, filter } = queryParams;

    return (
      <Query
        query={SilencesContent.query}
        fetchPolicy="cache-and-network"
        pollInterval={pollInterval}
        variables={{ ...match.params, limit, offset, order, filter }}
      >
        {({
          data: { environment } = {},
          loading,
          aborted,
          refetch,
          isPolling,
          startPolling,
          stopPolling,
        }) => {
          if (!environment && !loading && !aborted) {
            return <NotFoundView />;
          }

          return (
            <div>
              <ListToolbar
                renderSearch={
                  <SearchBox
                    placeholder="Filter silences…"
                    initialValue={filter}
                    onSearch={value => setQueryParams({ filter: value })}
                  />
                }
                renderMenuItems={
                  <React.Fragment>
                    <ModalController
                      renderModal={({ close }) => (
                        <SilenceEntryDialog
                          values={{
                            props: {},
                            ns: match.params,
                          }}
                          onClose={() => {
                            // TODO: Only refetch / poison list on success
                            refetch();
                            close();
                          }}
                        />
                      )}
                    >
                      {({ open }) => (
                        <CollapsingMenu.Button
                          title="New"
                          icon={<PlusIcon />}
                          onClick={() => open()}
                        />
                      )}
                    </ModalController>
                    <CollapsingMenu.Button
                      title="LIVE"
                      icon={<LiveIcon active={isPolling} />}
                      onClick={() =>
                        isPolling ? stopPolling() : startPolling(pollInterval)
                      }
                    />
                  </React.Fragment>
                }
              />
              <AppLayout.MobileFullWidthContent>
                <SilencesList
                  limit={limit}
                  offset={offset}
                  onChangeQuery={setQueryParams}
                  environment={environment}
                  loading={(loading && (!environment || !isPolling)) || aborted}
                  refetch={refetch}
                />
              </AppLayout.MobileFullWidthContent>
            </div>
          );
        }}
      </Query>
    );
  }
}

export default withQueryParams(["filter", "order", "offset", "limit"])(
  SilencesContent,
);
