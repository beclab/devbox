package envoy

import (
	"encoding/json"
	"fmt"
	"strings"

	envoy_config_bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_authz_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	envoy_router_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	originaldstv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/original_dst/v3"
	http_connection_manager_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/golang/protobuf/ptypes/duration"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

type ConfigBuilder struct {
	containers []*DevcontainerEndpoint
	websocket  bool
	owner      string
}

func (cb *ConfigBuilder) WithDevcontainers(containers []*DevcontainerEndpoint) *ConfigBuilder {
	cb.containers = containers
	return cb
}

func (cb *ConfigBuilder) WithWebsocket() *ConfigBuilder {
	cb.websocket = true
	return cb
}

func (cb *ConfigBuilder) Websocket() bool {
	return cb.websocket
}

func (cb *ConfigBuilder) Build() (string, error) {
	var bootstrap envoy_config_bootstrap.Bootstrap

	routes := []*routev3.Route{}

	for _, c := range cb.containers {
		if c.Port <= 0 || c.Port == 5000 {
			continue
		}
		routes = append(routes, &routev3.Route{
			Match: &routev3.RouteMatch{
				PathSpecifier: &routev3.RouteMatch_Prefix{
					Prefix: "/",
				},
				Headers: []*routev3.HeaderMatcher{
					{
						Name: ":authority",
						HeaderMatchSpecifier: &routev3.HeaderMatcher_SafeRegexMatch{
							SafeRegexMatch: &matcherv3.RegexMatcher{
								EngineType: &matcherv3.RegexMatcher_GoogleRe2{GoogleRe2: &matcherv3.RegexMatcher_GoogleRE2{}},
								Regex:      fmt.Sprintf("^[^.]+-%d\\.[^.]+\\..*$", c.Port),
							},
						},
					},
				},
			},
			Action: &routev3.Route_Route{
				Route: &routev3.RouteAction{
					ClusterSpecifier: &routev3.RouteAction_Cluster{
						Cluster: c.Name,
					},
					Timeout: &duration.Duration{
						Seconds: 300,
					},
				},
			},
		})
	}

	for _, c := range cb.containers {
		routes = append(routes, &routev3.Route{
			Match: &routev3.RouteMatch{
				PathSpecifier: &routev3.RouteMatch_Prefix{
					Prefix: c.Path,
				},
			},
			Action: &routev3.Route_Route{
				Route: &routev3.RouteAction{
					ClusterSpecifier: &routev3.RouteAction_Cluster{
						Cluster: c.Name,
					},
					Timeout: &duration.Duration{
						Seconds: 300,
					},
				},
			},
		})
	}

	//routes = append(routes,
	//	&routev3.Route{
	//		Match: &routev3.RouteMatch{
	//			PathSpecifier: &routev3.RouteMatch_Prefix{
	//				Prefix: "/",
	//			},
	//		},
	//		Action: &routev3.Route_Route{
	//			Route: &routev3.RouteAction{
	//				ClusterSpecifier: &routev3.RouteAction_Cluster{
	//					Cluster: "original_dst",
	//				},
	//				Timeout: &duration.Duration{
	//					Seconds: 300,
	//				},
	//			},
	//		},
	//	},
	//)

	lisenters := []*listenerv3.Listener{
		{
			Name: "devcontainer_proxy",

			//	  address:
			//		socket_address:
			//		  address: 0.0.0.0
			//		  port_value: 15003
			Address: &corev3.Address{
				Address: &corev3.Address_SocketAddress{
					SocketAddress: &corev3.SocketAddress{
						Address: "0.0.0.0",
						PortSpecifier: &corev3.SocketAddress_PortValue{
							PortValue: 15003,
						},
					},
				},
			},

			//	  listener_filters:
			//		- name: envoy.filters.listener.original_dst
			//		  typed_config:
			//			"@type": type.googleapis.com/envoy.extensions.filters.listener.original_dst.v3.OriginalDst
			ListenerFilters: []*listenerv3.ListenerFilter{
				{
					Name: "envoy.filters.listener.original_dst",
					ConfigType: &listenerv3.ListenerFilter_TypedConfig{
						TypedConfig: MessageToAny(&originaldstv3.OriginalDst{}),
					},
				},
			},

			//	  filter_chains:
			FilterChains: []*listenerv3.FilterChain{
				{
					Filters: []*listenerv3.Filter{
						{
							Name: "envoy.filters.network.http_connection_manager",
							ConfigType: &listenerv3.Filter_TypedConfig{
								TypedConfig: MessageToAny(&http_connection_manager_v3.HttpConnectionManager{
									StatPrefix: "dev-container",
									UpgradeConfigs: []*http_connection_manager_v3.HttpConnectionManager_UpgradeConfig{
										{
											UpgradeType: "websocket",
										},
									},
									SkipXffAppend: false,
									CodecType:     http_connection_manager_v3.HttpConnectionManager_AUTO,
									RouteSpecifier: &http_connection_manager_v3.HttpConnectionManager_RouteConfig{
										RouteConfig: &routev3.RouteConfiguration{
											Name: "local_route",
											VirtualHosts: []*routev3.VirtualHost{
												{
													Name:    "service",
													Domains: []string{"*"},
													Routes:  routes,
												},
											},
										},
									},
									HttpFilters: []*http_connection_manager_v3.HttpFilter{
										authFilter(cb.owner),
										{
											Name: "envoy.filters.http.router",
											ConfigType: &http_connection_manager_v3.HttpFilter_TypedConfig{
												TypedConfig: MessageToAny(&envoy_router_v3.Router{}),
											},
										},
									},

									//				http_protocol_options:
									//				  accept_http_10: true
									HttpProtocolOptions: &corev3.Http1ProtocolOptions{
										AcceptHttp_10: true,
									},
								}),
							},
						},
					},
				},
			},
		},
	}

	clusters := []*clusterv3.Cluster{
		{
			Name: "original_dst",
			ClusterDiscoveryType: &clusterv3.Cluster_Type{
				Type: clusterv3.Cluster_ORIGINAL_DST,
			},
			ConnectTimeout: &duration.Duration{
				Seconds: 5,
			},
			LbPolicy: clusterv3.Cluster_CLUSTER_PROVIDED,
		},

		{
			Name: "authelia",
			ClusterDiscoveryType: &clusterv3.Cluster_Type{
				Type: clusterv3.Cluster_LOGICAL_DNS,
			},
			ConnectTimeout: &duration.Duration{
				Seconds: 1,
			},
			DnsRefreshRate: &duration.Duration{
				Seconds: 600,
			},
			DnsLookupFamily: clusterv3.Cluster_V4_ONLY,
			LbPolicy:        clusterv3.Cluster_ROUND_ROBIN,
			LoadAssignment: &endpointv3.ClusterLoadAssignment{
				ClusterName: "authelia",
				Endpoints: []*endpointv3.LocalityLbEndpoints{
					{
						LbEndpoints: []*endpointv3.LbEndpoint{
							{
								HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
									Endpoint: &endpointv3.Endpoint{
										Address: &corev3.Address{
											Address: &corev3.Address_SocketAddress{
												SocketAddress: &corev3.SocketAddress{
													Address: fmt.Sprintf("authelia-backend.user-system-%s", cb.owner),
													PortSpecifier: &corev3.SocketAddress_PortValue{
														PortValue: 9091,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if cb.websocket {
		clusters = append(clusters, &clusterv3.Cluster{
			Name: "ws_gateway",
			ClusterDiscoveryType: &clusterv3.Cluster_Type{
				Type: clusterv3.Cluster_LOGICAL_DNS,
			},
			ConnectTimeout: &duration.Duration{
				Seconds: 5,
			},
			DnsRefreshRate: &duration.Duration{
				Seconds: 600,
			},
			DnsLookupFamily: clusterv3.Cluster_V4_ONLY,
			LbPolicy:        clusterv3.Cluster_ROUND_ROBIN,
			LoadAssignment: &endpointv3.ClusterLoadAssignment{
				ClusterName: "ws_gateway",
				Endpoints: []*endpointv3.LocalityLbEndpoints{
					{
						LbEndpoints: []*endpointv3.LbEndpoint{
							{
								HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
									Endpoint: &endpointv3.Endpoint{
										Address: &corev3.Address{
											Address: &corev3.Address_SocketAddress{
												SocketAddress: &corev3.SocketAddress{
													Address: "localhost",
													PortSpecifier: &corev3.SocketAddress_PortValue{
														PortValue: 40010,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})
	}

	for _, c := range cb.containers {
		// add route to http listener

		// add cluster to devcontainer endpoint
		// - name: dev
		//   connect_timeout: 5s
		//   type: LOGICAL_DNS
		//   dns_lookup_family: V4_ONLY
		//   dns_refresh_rate: 600s
		//   lb_policy: ROUND_ROBIN
		//   load_assignment:
		// 	   cluster_name: dev
		// 	   endpoints:
		// 	   - lb_endpoints:
		// 		  - endpoint:
		// 			  address:
		// 				socket_address:
		// 				  address: localhost
		// 				  port_value: 5000
		clusters = append(clusters, &clusterv3.Cluster{
			Name: c.Name,
			ConnectTimeout: &duration.Duration{
				Seconds: 5,
			},
			DnsRefreshRate: &duration.Duration{
				Seconds: 600,
			},
			DnsLookupFamily: clusterv3.Cluster_V4_ONLY,
			LbPolicy:        clusterv3.Cluster_ROUND_ROBIN,
			ClusterDiscoveryType: &clusterv3.Cluster_Type{
				Type: clusterv3.Cluster_LOGICAL_DNS,
			},
			LoadAssignment: &endpointv3.ClusterLoadAssignment{
				ClusterName: c.Name,
				Endpoints: []*endpointv3.LocalityLbEndpoints{
					{
						LbEndpoints: []*endpointv3.LbEndpoint{
							{
								HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
									Endpoint: &endpointv3.Endpoint{
										Address: &corev3.Address{
											Address: &corev3.Address_SocketAddress{
												SocketAddress: &corev3.SocketAddress{
													Address: c.Host,
													PortSpecifier: &corev3.SocketAddress_PortValue{
														PortValue: uint32(c.Port),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})
	}

	bootstrap.StaticResources = &envoy_config_bootstrap.Bootstrap_StaticResources{
		Listeners: lisenters,
		Clusters:  clusters,
	}

	m, err := ToJSONMap(&bootstrap)
	if err != nil {
		klog.Error("ToJSONMap ", err)
		return "", err
	}

	mBytes, err := json.Marshal(SnakeCaseMarshaller{Value: m})
	if err != nil {
		klog.Error("SnakeCaseMarshaller ", err)
		return "", err
	}

	config, err := yaml.JSONToYAML(mBytes)
	if err != nil {
		klog.Error("JSONToYAML: ", err)
	}

	cfgStr := strings.ReplaceAll(string(config), "google_re_2:", "google_re2:")
	return cfgStr, err
}

func authFilter(owner string) *http_connection_manager_v3.HttpFilter {
	return &http_connection_manager_v3.HttpFilter{
		Name: "envoy.filters.http.ext_authz",
		ConfigType: &http_connection_manager_v3.HttpFilter_TypedConfig{
			TypedConfig: MessageToAny(&envoy_authz_v3.ExtAuthz{
				Services: &envoy_authz_v3.ExtAuthz_HttpService{
					HttpService: &envoy_authz_v3.HttpService{
						PathPrefix: "/api/verify/",
						ServerUri: &corev3.HttpUri{
							Uri: fmt.Sprintf("authelia-backend.user-system-%s:9091", owner),
							HttpUpstreamType: &corev3.HttpUri_Cluster{
								Cluster: "authelia",
							},
							Timeout: &duration.Duration{
								Seconds: 0,
								Nanos:   250000000,
							},
						},
						AuthorizationRequest: &envoy_authz_v3.AuthorizationRequest{
							AllowedHeaders: &matcherv3.ListStringMatcher{
								Patterns: []*matcherv3.StringMatcher{
									{
										MatchPattern: &matcherv3.StringMatcher_Exact{
											Exact: "accept",
										},
									},
									{
										MatchPattern: &matcherv3.StringMatcher_Exact{
											Exact: "cookie",
										},
									},
									{
										MatchPattern: &matcherv3.StringMatcher_Exact{
											Exact: "proxy-authorization",
										},
									},
									{
										MatchPattern: &matcherv3.StringMatcher_Prefix{
											Prefix: "x-unauth-",
										},
									},
									{
										MatchPattern: &matcherv3.StringMatcher_Exact{
											Exact: "x-authorization",
										},
									},
									{
										MatchPattern: &matcherv3.StringMatcher_Exact{
											Exact: "x-bfl-user",
										},
									},
									{
										MatchPattern: &matcherv3.StringMatcher_Exact{
											Exact: "terminus-nonce",
										},
									},
								},
							},
							HeadersToAdd: []*corev3.HeaderValue{
								{
									Key:   "X-Forwarded-Method",
									Value: "%REQ(:METHOD)%",
								},
								{
									Key:   "X-Forwarded-Proto",
									Value: "%REQ(:SCHEME)%",
								},
								{
									Key:   "X-Forwarded-Host",
									Value: "%REQ(:AUTHORITY)%",
								},
								{
									Key:   "X-Forwarded-Uri",
									Value: "%REQ(:PATH)%",
								},
								{
									Key:   "X-Forwarded-For",
									Value: "%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%",
								},
							},
						},
						AuthorizationResponse: &envoy_authz_v3.AuthorizationResponse{
							AllowedUpstreamHeaders: &matcherv3.ListStringMatcher{
								Patterns: []*matcherv3.StringMatcher{
									{
										MatchPattern: &matcherv3.StringMatcher_Exact{
											Exact: "authorization",
										},
									},
									{
										MatchPattern: &matcherv3.StringMatcher_Exact{
											Exact: "proxy-authorization",
										},
									},
									{
										MatchPattern: &matcherv3.StringMatcher_Prefix{
											Prefix: "remote-",
										},
									},
									{
										MatchPattern: &matcherv3.StringMatcher_Prefix{
											Prefix: "authelia-",
										},
									},
								},
							},
							AllowedClientHeaders: &matcherv3.ListStringMatcher{
								Patterns: []*matcherv3.StringMatcher{
									{
										MatchPattern: &matcherv3.StringMatcher_Exact{
											Exact: "set-cookie",
										},
									},
								},
							},
						},
					},
				},
				FailureModeAllow: false,
			}),
		},
	}
}
