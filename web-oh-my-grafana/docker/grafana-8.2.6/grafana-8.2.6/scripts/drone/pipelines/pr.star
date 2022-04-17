load(
    'scripts/drone/steps/lib.star',
    'lint_backend_step',
    'codespell_step',
    'shellcheck_step',
    'build_backend_step',
    'build_frontend_step',
    'build_plugins_step',
    'test_backend_step',
    'test_frontend_step',
    'gen_version_step',
    'package_step',
    'e2e_tests_server_step',
    'e2e_tests_step',
    'build_storybook_step',
    'build_frontend_docs_step',
    'build_docs_website_step',
    'copy_packages_for_docker_step',
    'build_docker_images_step',
    'postgres_integration_tests_step',
    'mysql_integration_tests_step',
    'redis_integration_tests_step',
    'memcached_integration_tests_step',
    'benchmark_ldap_step',
    'validate_scuemata_step',
    'test_a11y_frontend_step_pr',
)

load(
    'scripts/drone/services/services.star',
    'integration_test_services',
    'ldap_service',
)

load(
    'scripts/drone/utils/utils.star',
    'pipeline',
)

ver_mode = 'pr'

def pr_pipelines(edition):
    services = integration_test_services(edition)
    variants = ['linux-x64', 'linux-x64-musl', 'osx64', 'win64', 'armv6',]
    include_enterprise2 = edition == 'enterprise'
    steps = [
        codespell_step(),
        shellcheck_step(),
        lint_backend_step(edition=edition),
        test_backend_step(edition=edition),
        test_frontend_step(),
        build_backend_step(edition=edition, ver_mode=ver_mode, variants=variants),
        build_frontend_step(edition=edition, ver_mode=ver_mode),
        build_plugins_step(edition=edition),
        validate_scuemata_step(),
    ]

    # Have to insert Enterprise2 steps before they're depended on (in the gen-version step)
    if include_enterprise2:
        edition2 = 'enterprise2'
        steps.append(benchmark_ldap_step())
        services.append(ldap_service())
        steps.extend([
            lint_backend_step(edition=edition2),
            test_backend_step(edition=edition2),
            build_backend_step(edition=edition2, ver_mode=ver_mode, variants=['linux-x64']),
        ])

    # Insert remaining steps
    steps.extend([
        gen_version_step(ver_mode=ver_mode, include_enterprise2=include_enterprise2),
        package_step(edition=edition, ver_mode=ver_mode, variants=variants),
        e2e_tests_server_step(edition=edition),
        e2e_tests_step(edition=edition),
        build_storybook_step(edition=edition, ver_mode=ver_mode),
        test_a11y_frontend_step_pr(edition=edition),
        build_frontend_docs_step(edition=edition),
        build_docs_website_step(),
        copy_packages_for_docker_step(),
        build_docker_images_step(edition=edition, ver_mode=ver_mode, archs=['amd64',]),
        postgres_integration_tests_step(),
        mysql_integration_tests_step(),
    ])

    if include_enterprise2:
        steps.extend([
            redis_integration_tests_step(),
            memcached_integration_tests_step(),
            package_step(edition=edition2, ver_mode=ver_mode, variants=['linux-x64']),
            e2e_tests_server_step(edition=edition2, port=3002),
            e2e_tests_step(edition=edition2, port=3002),
        ])

    trigger = {
        'event': ['pull_request',],
    }
    return [
        pipeline(
            name='test-pr', edition=edition, trigger=trigger, services=services, steps=steps,
            ver_mode=ver_mode,
        ),
    ]
