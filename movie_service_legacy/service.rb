require 'json'
require 'sinatra'
require 'logger'
require 'aphorism'
require 'color-generator'

class MovieService < Sinatra::Base
  configure do
    set :show_exceptions, false
    set :logger, Logger.new("sinatra.log", 'daily')
    enable :logging, :dump_errors
    set :raise_errors, true
  end

  before do
    @generator = ColorGenerator.new(saturation: 0.5, lightness: 0.5)
  end

  get '/movies' do
    content_type :json
    read_data.to_json
  end

  get '/movies/:id' do
    movies = read_data
    movie = movies.find { |m| m["imdbID"] == params.fetch(:id, '') }

    if movie
      content_type :json
      movie.to_json
    else
      status 404
      { error: "Movie with requested ID not found" }.to_json
    end
  end

  get '/' do
    color = @generator.create_hex
    orator = Aphorism::Orator.new
    message = orator.say.split("-")
    "<p style=\"text-align: center; font-size: 2em; color: #{color}\">
      #{message[0]} <br> - <span style=\"font-style: italic;\">#{message[1]}</span>
    </p>"
  end

  not_found do
    status 404
    { error: "There is nothing to do here! 404!" }.to_json
  end

  error 500 do
    status 500
    { error: "Something went wrong!" }.to_json
  end

  private

  def read_data
    JSON.parse(File.read("movies.json"))
  rescue StandardError => e
    logger.error "Failed to read movies.json: #{e.message}"
    []
  end
end
